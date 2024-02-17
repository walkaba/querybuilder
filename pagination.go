package querybuilder

import "fmt"

type IPaginationStrategy interface {
	First(map[string]int) string
	Last(map[string]int, int) string
	Next(map[string]int) string
	Prev(map[string]int) string
}

type PageSizeStrategy struct{}

func (ps PageSizeStrategy) First(c map[string]int) string {
	var (
		p int
		s int
	)
	if size, ok := c["size"]; ok {
		s = size
	} else {
		return ""
	}
	p = 0
	return fmt.Sprintf("page[size]=%d&page[page]=%d", s, p)
}

func (os PageSizeStrategy) Last(c map[string]int, total int) string {
	var (
		p int
		s int
	)
	if size, ok := c["size"]; ok {
		s = size
	} else {
		return ""
	}
	p = total / s
	return fmt.Sprintf("page[size]=%d&page[page]=%d", s, p)
}

func (ps PageSizeStrategy) Next(c map[string]int) string {
	var (
		p int
		s int
	)
	if size, ok := c["size"]; ok {
		s = size
	} else {
		return ""
	}
	if page, ok := c["page"]; ok {
		p = page + 1
	} else {
		p = 0
	}
	return fmt.Sprintf("page[size]=%d&page[page]=%d", s, p)
}

func (ps PageSizeStrategy) Prev(c map[string]int) string {
	var (
		p int
		s int
	)
	if size, ok := c["size"]; ok {
		s = size
	} else {
		return ""
	}
	if page, ok := c["page"]; ok {
		p = page - 1
	} else {
		p = 0
	}
	if p < 0 {
		p = 0
	}
	return fmt.Sprintf("page[size]=%d&page[page]=%d", s, p)
}
