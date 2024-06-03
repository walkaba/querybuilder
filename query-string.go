package querybuilder

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if _, ok := options.Page["size"]; ok {
		options.SetPaginationStrategy(&PageSizeStrategy{})
	}
	return options, nil
}

func parseFields(qs *string) []string {
	var fields []string
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
	cords := re.FindStringIndex(*qs)
	var r string
	if cords != nil {
		if cords[0] == 0 {
			r = (*qs)[cords[0]:cords[1]]
			*qs = (*qs)[cords[1]:]
			return r
		}
		if cords[1] == len(*qs) {
			r = (*qs)[cords[0]:cords[1]]
			*qs = (*qs)[0:cords[0]]
			return r
		}
		r = (*qs)[cords[0]:cords[1]]
		*qs = fmt.Sprintf("%s%s", (*qs)[0:cords[0]], (*qs)[cords[1]:])
		return r
	}
	return r
}

func parseSort(qs *string) []string {
	var sort []string
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

func isUnableToParse(terms [][]string, values [][]string) error {
	if len(terms) > 0 && len(terms) > len(values) {
		return errors.New("unable to parse: an object hierarchy has been provided")
	}
	return nil
}

func parseBracketParams(qs string, o *Options) error {
	o.Filter = map[string]interface{}{}
	o.Page = map[string]int{}
	terms := bracketRE.FindAllStringSubmatch(qs, -1)
	values := bracketValueRE.FindAllStringSubmatch(qs, -1)
	err := isUnableToParse(terms, values)
	if err != nil {
		return err
	}
	if o.Filter == nil {
		o.Filter = map[string]interface{}{}
	}
	if o.Page == nil {
		o.Page = map[string]int{}
	}
	for i, term := range terms {
		switch strings.ToLower(term[1]) {
		case "filter":
			err := SetJSONValue(term[2], values[i][1], o.Filter)
			if err != nil {
				return err
			}
		case "page":
			v, err := strconv.ParseInt(values[i][1], 0, 64)
			if err != nil {
				return err
			}
			o.Page[term[2]] = int(v)
		}
	}
	return nil
}

func validateValue(value interface{}) interface{} {
	str := value.(string)
	if valueInterger, err := strconv.Atoi(str); err == nil {
		return valueInterger
	} else if valueFloat, err := strconv.ParseFloat(str, 64); err == nil {
		return valueFloat
	} else {
		return value
	}
}

func setValueInMapOrArray(current interface{}, keys []string, value interface{}) interface{} {
	if len(keys) == 0 {
		strVal, ok := value.(string)
		if ok && strings.Contains(strVal, ",") {
			return strings.Split(strVal, ",")
		}
		return validateValue(value)
	}

	key := keys[0]
	remainingKeys := keys[1:]

	index, err := strconv.Atoi(key)
	if err == nil {
		var array []interface{}
		if current != nil {
			array = current.([]interface{})
		}

		if len(array) <= index {
			newArray := make([]interface{}, index+1)
			copy(newArray, array)
			array = newArray
		}

		array[index] = setValueInMapOrArray(array[index], remainingKeys, value)
		return array
	} else {
		var m map[string]interface{}
		if current != nil {
			m = current.(map[string]interface{})
		} else {
			m = make(map[string]interface{})
		}

		m[key] = setValueInMapOrArray(m[key], remainingKeys, value)
		return m
	}
}

func SetJSONValue(path string, value interface{}, filter map[string]interface{}) error {
	if strings.HasPrefix(path, "$or][") {
		keys := strings.Split(path, "][")
		result := setValueInMapOrArray(filter, keys, value)
		for k, v := range result.(map[string]interface{}) {
			filter[k] = v
		}
		return nil
	} else {
		if commaRE.MatchString(value.(string)) {
			filter[path] = commaRE.Split(value.(string), -1)
			return nil
		}
		filter[path] = []string{value.(string)}
		return nil
	}
}

type QueryBuilder struct {
	fieldTypes       map[string]string
	strictValidation bool
}

func NewQueryBuilder(strictValidation ...bool) *QueryBuilder {
	qb := QueryBuilder{
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
	if len(fields) == 0 {
		return nil
	}
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
				return fmt.Errorf("field %s does not exist in collection", field)
			}
		}
		prj[field] = val
	}
	if len(prj) > 0 {
		opts.SetProjection(prj)
	}
	return nil
}

func (qb QueryBuilder) setSortOptions(fields []string, opts *options.FindOptions) error {
	if len(fields) == 0 {
		return nil
	}
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
				return fmt.Errorf("field %s does not exist in collection", field)
			}
		}
		sort[field] = val
	}
	opts.SetSort(sort)
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
	filters := parseFilters(opt.Filter)
	return filters, nil

}

func parseFilters(filter map[string]interface{}) bson.D {
	var filters bson.D
	for k, v := range filter {
		switch v := v.(type) {
		case []interface{}:
			subParts, _ := processArray(k, v)
			filters = append(filters, bson.E{Key: k, Value: subParts})
			continue
		case []string:
			subParts := checkFilter(k, v)
			filters = append(filters, subParts...)
			continue
		case string:
			subParts := checkFilter(k, []string{v})
			filters = append(filters, subParts...)
			continue
		default:
			subParts, _ := processMap(k, v)
			for key, value := range subParts {
				filters = append(filters, bson.E{Key: key, Value: value})
			}
			continue
		}
	}
	return filters
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
	case "<=>":
		return "like"
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
					Key:   field,
					Value: acc,
				}}
			}

			var includes []string
			for _, value := range values {
				includes = append(includes, value)
			}
			return bson.D{{
				Key: field,
				Value: bson.D{{
					Key: "$in", Value: includes,
				}},
			}}
		}
		if field == "_id" {
			res, _ := primitive.ObjectIDFromHex(values[0])
			return bson.D{{
				Key: field,
				Value: bson.D{{
					Key: "$eq", Value: res,
				}},
			}}
		}
		check := checkConstraints(values[0])
		if check == "like" {
			return bson.D{{
				Key: field,
				Value: bson.D{{
					Key: "$regex", Value: values[0],
				}},
			}}
		}
		return bson.D{{
			Key: field,
			Value: bson.D{{
				Key:   check,
				Value: values[0],
			}},
		}}
	} else {
		if len(values) > 1 {
			check := checkConstraints(values[0])
			var acc bson.D
			isNull := false
			for _, value := range values {
				if value != "null" {
					acc = append(acc,
						bson.E{
							Key:   check,
							Value: value,
						},
					)
				} else {
					isNull = true
				}
			}

			if isNull {
				return bson.D{
					{
						Key: "$or",
						Value: bson.A{
							bson.D{{
								Key: key[0],
								Value: bson.M{
									"$elemMatch": bson.M{
										key[1]: acc}}},
							},
							bson.D{
								{Key: key[0], Value: bson.M{"$size": 0}},
							},
						},
					},
				}
			}
			return bson.D{{
				Key: key[0], Value: bson.D{{
					Key: "$elemMatch", Value: bson.D{{
						Key: key[1], Value: acc,
					}},
				}},
			}}
		}

		return bson.D{{
			Key: key[0], Value: bson.D{{
				Key: "$elemMatch", Value: bson.D{{
					Key: key[1], Value: values[0],
				}},
			}},
		}}
	}
}

func processArray(operator string, value interface{}) ([]interface{}, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid format for %s", operator)
	}
	var parts []interface{}
	for _, v := range values {
		subPart := primitive.D{}
		if result, ok := v.(map[string]interface{}); ok {
			subPart = parseFilters(result)
		}
		parts = append(parts, subPart)
	}
	return parts, nil
}

func processMap(key string, value interface{}) (map[string]interface{}, error) {
	subMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid format for %s", key)
	}
	result := make(map[string]interface{})
	hasReservedKey := false
	for subKey, subValue := range subMap {
		if isReservedKey(subKey) {
			hasReservedKey = true
		}
		switch subKey {
		case "$in":
			result[subKey] = formatArray(subValue)
			continue
		case "$size", "$not", "$lte", "$gte", "$ne", "$lt", "$gt", "$eq":
			result[subKey] = subValue
			continue
		case "$like":
			result["$regex"] = subValue
			continue
		default:
			result[subKey] = formatValue(subValue)
			continue
		}
	}
	if len(result) > 0 && !hasReservedKey {
		return map[string]interface{}{key: map[string]interface{}{"$elemMatch": result}}, nil
	}
	return map[string]interface{}{key: result}, nil
}

func isReservedKey(key string) bool {
	reservedKeys := map[string]bool{
		"$in":   true,
		"$size": true,
		"$or":   true,
		"$not":  true,
		"$lte":  true,
		"$gte":  true,
		"$ne":   true,
		"$lt":   true,
		"$gt":   true,
		"$like": true,
		"$eq":   true,
	}
	return reservedKeys[key]
}

func formatArray(value interface{}) []interface{} {
	values, ok := value.([]interface{})
	if !ok {
		return []interface{}{}
	}
	var parts []interface{}
	for _, v := range values {
		parts = append(parts, formatValue(v))
	}
	return parts
}

func formatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if v == "null" {
			return nil
		}
		return v
	case float64, int, bool:
		return v
	default:
		return v
	}
}
