package querybuilder

import (
	"fmt"
	"strings"
)

type Options struct {
	ps IPaginationStrategy
	qs string

	Fields []string               `json:"fields,omitempty"`
	Filter map[string]interface{} `json:"filter,omitempty"`
	Page   map[string]int         `json:"page"`
	Sort   []string               `json:"sort,omitempty"`
}

func (o Options) ContainsFilterField(field string) bool {
	fields := []string{}
	for f := range o.Filter {
		fields = append(fields, f)
	}

	return contains(fields, field, false)
}

func (o Options) ContainsSortField(field string) bool {
	return contains(o.Sort, field, true)
}

func (o Options) First() string {
	if len(o.Page) == 0 || o.ps == nil {
		return buildQuerystring(o.Filter, o.Fields, "", o.Sort)
	}
	po := o.ps.First(o.Page)
	qs := buildQuerystring(o.Filter, o.Fields, po, o.Sort)
	return qs
}

func (o Options) Last(total int) string {
	if len(o.Page) == 0 || o.ps == nil {
		return buildQuerystring(o.Filter, o.Fields, "", o.Sort)
	}
	po := o.ps.Last(o.Page, total)
	qs := buildQuerystring(o.Filter, o.Fields, po, o.Sort)
	return qs
}

func (o Options) Next() string {
	if len(o.Page) == 0 || o.ps == nil {
		return buildQuerystring(o.Filter, o.Fields, "", o.Sort)
	}
	po := o.ps.Next(o.Page)
	qs := buildQuerystring(o.Filter, o.Fields, po, o.Sort)
	return qs
}

func (o Options) PaginationStrategy() IPaginationStrategy {
	return o.ps
}

func (o Options) Prev() string {
	if len(o.Page) == 0 || o.ps == nil {
		return buildQuerystring(o.Filter, o.Fields, "", o.Sort)
	}
	po := o.ps.Prev(o.Page)
	qs := buildQuerystring(o.Filter, o.Fields, po, o.Sort)
	return qs
}

func (o *Options) SetPaginationStrategy(ps IPaginationStrategy) {
	o.ps = ps
}

func buildFilterQuery(b *strings.Builder, filter map[string]interface{}, prefix string) {
	ra := false

	for key, value := range filter {
		if ra {
			fmt.Fprint(b, "&")
		}
		switch v := value.(type) {
		case map[string]interface{}:
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "[" + key + "]"
			}
			buildFilterQuery(b, v, newPrefix)
		case []interface{}:
			for i, subValue := range v {
				newPrefix := fmt.Sprintf("%s[%d]", key, i)
				if prefix != "" {
					newPrefix = prefix + "[" + newPrefix + "]"
				}
				if subMap, ok := subValue.(map[string]interface{}); ok {
					buildFilterQuery(b, subMap, newPrefix)
				} else {
					fmt.Fprintf(b, "filter[%s]=", newPrefix)
					fmt.Fprint(b, subValue)
				}
			}
		default:
			fmt.Fprintf(b, "filter[%s]=", prefix+key)
			fmt.Fprint(b, v)
		}
		ra = true
	}
}

func buildQuerystring(filter map[string]interface{}, fields []string, page string, sort []string) string {
	b := strings.Builder{}
	ra := false

	buildFilterQuery(&b, filter, "")

	if len(fields) > 0 {
		if ra {
			fmt.Fprint(&b, "&")
		}
		fmt.Fprintf(&b, "fields=")
		for i, field := range fields {
			if i > 0 {
				fmt.Fprint(&b, ",")
			}
			fmt.Fprint(&b, field)
		}
		ra = true
	}
	if page != "" {
		if ra {
			fmt.Fprint(&b, "&")
		}
		fmt.Fprint(&b, page)
		ra = true
	}
	if len(sort) > 0 {
		if ra {
			fmt.Fprint(&b, "&")
		}
		fmt.Fprintf(&b, "sort=")
		for i, field := range sort {
			if i > 0 {
				fmt.Fprint(&b, ",")
			}
			fmt.Fprint(&b, field)
		}
	}
	return b.String()
}

func contains(list []string, value string, stripPrefix bool) bool {
	if len(list) == 0 {
		return false
	}
	if value == "" {
		return false
	}
	for _, search := range list {
		if stripPrefix {
			if len(search) > 2 && (search[0:2] == "<=" || search[0:2] == ">=" || search[0:2] == "!=") {
				search = search[2:]
			}
			if len(search) > 1 && (search[0:1] == "<" || search[0:1] == ">" || search[0:1] == "-" || search[0:1] == "+" || search[0:1] == "!") {
				search = search[1:]
			}
		}
		if search == value {
			return true
		}
	}
	return false
}
