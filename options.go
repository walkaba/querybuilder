package querybuilder

import (
	"fmt"
	"strings"
)

type Options struct {
	ps IPaginationStrategy
	qs string

	Fields []string            `json:"fields,omitempty"`
	Filter map[string][]string `json:"filter,omitempty"`
	Page   map[string]int      `json:"page"`
	Sort   []string            `json:"sort,omitempty"`
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

func buildQuerystring(filter map[string][]string, fields []string, page string, sort []string) string {
	b := strings.Builder{}
	ra := false
	for field, filter := range filter {
		if ra {
			fmt.Fprint(&b, "&")
		}
		ra = true
		fmt.Fprintf(&b, "filter[%s]=", field)
		for i, value := range filter {
			if i > 0 {
				fmt.Fprint(&b, ",")
			}
			fmt.Fprint(&b, value)
		}
	}
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
	}
	if page != "" {
		if ra {
			fmt.Fprint(&b, "&")
		}
		ra = true
		fmt.Fprint(&b, page)
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
