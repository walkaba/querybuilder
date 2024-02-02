package querybuilder

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	bracketRE      = regexp.MustCompile(`(?P<typ>filter|sort|page)\[([^&]+?)\](\={1})`)
	bracketValueRE = regexp.MustCompile(`\]\=(.*?)(\&|\z)`)
	commaRE        = regexp.MustCompile(`\s?\,\s?`)
	fieldsRE       = regexp.MustCompile(`fields=(?P<field>.+?)(\&|\z)`)
	sortRE         = regexp.MustCompile(`sort=(?P<field>.+?)(\&|\z)`)
)

func FromQueryString(qs string) (Options, error) {
	if qs == "" {
		return Options{}, nil
	}
	uqs, err := url.QueryUnescape(qs)
	if err != nil {
		return Options{}, err
	}
	options := Options{
		qs: uqs,
	}
	options.Fields = parseFields(&uqs)
	options.Sort = parseSort(&uqs)
	if err := parseBracketParams(uqs, &options); err != nil {
		return options, err
	}
	if _, ok := options.Page["limit"]; ok {
		options.SetPaginationStrategy(&OffsetStrategy{})
	}
	if _, ok := options.Page["size"]; ok {
		options.SetPaginationStrategy(&PageSizeStrategy{})
	}
	return options, nil
}

func parseFields(qs *string) []string {
	fields := []string{}
	fieldNames := fieldsRE.FindAllStringSubmatch(extract(qs, *fieldsRE), -1)
	for fieldNames != nil {
		for _, field := range fieldNames {
			if commaRE.MatchString(field[1]) {
				fa := commaRE.Split(field[1], -1)
				fields = append(fields, fa...)
				continue
			}

			fields = append(fields, field[1])
		}
		fieldNames = fieldsRE.FindAllStringSubmatch(extract(qs, *fieldsRE), -1)
	}

	return fields
}

func extract(qs *string, re regexp.Regexp) string {
	coords := re.FindStringIndex(*qs)
	var r string
	if coords != nil {
		if coords[0] == 0 {
			r = (*qs)[coords[0]:coords[1]]
			*qs = (*qs)[coords[1]:]
			return r
		}
		if coords[1] == len(*qs) {
			r = (*qs)[coords[0]:coords[1]]
			*qs = (*qs)[0:coords[0]]
			return r
		}
		r = (*qs)[coords[0]:coords[1]]
		*qs = fmt.Sprintf("%s%s", (*qs)[0:coords[0]], (*qs)[coords[1]:])
		return r
	}
	return r
}

func parseSort(qs *string) []string {
	sort := []string{}
	fieldNames := sortRE.FindAllStringSubmatch(extract(qs, *sortRE), -1)
	for fieldNames != nil {
		for _, field := range fieldNames {
			if commaRE.MatchString(field[1]) {
				fa := commaRE.Split(field[1], -1)
				sort = append(sort, fa...)
				continue
			}
			sort = append(sort, field[1])
		}
		fieldNames = sortRE.FindAllStringSubmatch(extract(qs, *sortRE), -1)
	}
	return sort
}

func parseBracketParams(qs string, o *Options) error {
	o.Filter = map[string][]string{}
	o.Page = map[string]int{}
	terms := bracketRE.FindAllStringSubmatch(qs, -1)
	values := bracketValueRE.FindAllStringSubmatch(qs, -1)
	if len(terms) > 0 && len(terms) > len(values) {
		return errors.New("unable to parse: an object hierarchy has been provided")
	}
	for i, term := range terms {
		switch strings.ToLower(term[1]) {
		case "filter":
			if o.Filter == nil {
				o.Filter = map[string][]string{}
			}
			if commaRE.MatchString(values[i][1]) {
				o.Filter[term[2]] = commaRE.Split(values[i][1], -1)
				continue
			}
			o.Filter[term[2]] = []string{values[i][1]}
		case "page":
			if o.Page == nil {
				o.Page = map[string]int{}
			}
			v, err := strconv.ParseInt(values[i][1], 0, 64)
			if err != nil {
				return err
			}
			o.Page[term[2]] = int(v)
		}
	}
	return nil
}

type QueryBuilder struct {
	collection       string
	fieldTypes       map[string]string
	strictValidation bool
}

func NewQueryBuilder(collection string, strictValidation ...bool) *QueryBuilder {
	qb := QueryBuilder{
		collection:       collection,
		fieldTypes:       map[string]string{},
		strictValidation: false,
	}
	if len(strictValidation) > 0 {
		qb.strictValidation = strictValidation[0]
	}
	return &qb
}

func (qb QueryBuilder) setPaginationOptions(pagination map[string]int, opts *options.FindOptions) {
	if limit, ok := pagination["limit"]; ok {
		opts.SetLimit(int64(limit))
		if offset, ok := pagination["offset"]; ok {
			opts.SetSkip(int64(offset))
		}
		if skip, ok := pagination["skip"]; ok {
			opts.SetSkip(int64(skip))
		}
	}
	if size, ok := pagination["size"]; ok {
		opts.SetLimit(int64(size))
		if page, ok := pagination["page"]; ok {
			opts.SetSkip(int64(page * size))
		}
	}
}

func (qb QueryBuilder) setProjectionOptions(fields []string, opts *options.FindOptions) error {
	if len(fields) > 0 {
		prj := map[string]int{}
		for _, field := range fields {
			val := 1
			if field[0:1] == "-" {
				field = field[1:]
				val = 0
			}
			if len(field) > 0 && field[0:1] == "+" {
				field = field[1:]
			}
			if qb.strictValidation {
				if _, ok := qb.fieldTypes[field]; !ok {
					// we have a problem
					return fmt.Errorf("field %s does not exist in collection %s", field, qb.collection)
				}
			}
			prj[field] = val
		}
		if len(prj) > 0 {
			opts.SetProjection(prj)
		}
	}

	return nil
}

func (qb QueryBuilder) setSortOptions(fields []string, opts *options.FindOptions) error {
	if len(fields) > 0 {
		sort := map[string]int{}
		for _, field := range fields {
			val := 1
			if field[0:1] == "-" {
				field = field[1:]
				val = -1
			}
			if field[0:1] == "+" {
				field = field[1:]
			}
			if qb.strictValidation {
				if _, ok := qb.fieldTypes[field]; !ok {
					return fmt.Errorf("field %s does not exist in collection %s", field, qb.collection)
				}
			}
			sort[field] = val
		}
		opts.SetSort(sort)
	}
	return nil
}

func (qb QueryBuilder) FindOptions(qo Options) (*options.FindOptions, error) {
	opts := options.Find()
	qb.setPaginationOptions(qo.Page, opts)
	if err := qb.setProjectionOptions(qo.Fields, opts); err != nil {
		return nil, err
	}
	if err := qb.setSortOptions(qo.Sort, opts); err != nil {
		return nil, err
	}
	return opts, nil
}

func (qb QueryBuilder) Filter(opt Options) (bson.D, error) {
	var filters bson.D
	for k, v := range opt.Filter {
		filters = checkFilter(k, v)
	}
	return filters, nil
}

func compareOperator(value string) string {
	switch value {
	case "<>":
		return "$not"
	case "<=":
		return "$lte"
	case ">=":
		return "$gte"
	case "!=":
		return "$ne"
	case "<":
		return "$lt"
	case ">":
		return "$gt"
	//case "<=>":
	//	return "like"
	default:
		return "$eq"
	}
}

func checkConstraints(values string) string {
	constraints := []string{"<>", "<=", ">=", "!=", "<", ">", "<=>"}
	for _, constr := range constraints {
		result := strings.Split(values, constr)
		if len(result) > 1 {
			return compareOperator(values)
		}
	}
	return "$eq"
}

func checkFilter(field string, values []string) bson.D {
	key := strings.Split(field, "][")
	if len(key) == 1 {
		if len(values) > 1 {
			check := checkConstraints(values[0])
			if check == "$lt" || check == "$gt" || check == "$gte" || check == "$lte" {
				var acc bson.D
				for _, value := range values {
					acc = append(acc,
						bson.E{
							Key:   check,
							Value: value,
						},
					)
				}
				return bson.D{{
					field,
					acc,
				}}
			}

			var includes []string
			for _, value := range values {
				includes = append(includes, value)
			}
			return bson.D{{
				field,
				bson.D{{
					"$in", includes,
				}},
			}}
		}
		return bson.D{{
			field,
			bson.D{{
				checkConstraints(values[0]),
				values[0],
			}},
		}}
	}

	return bson.D{{
		key[0], bson.D{{
			"$elemMatch", bson.D{{
				key[1], values[0],
			}},
		}},
	}}
}
