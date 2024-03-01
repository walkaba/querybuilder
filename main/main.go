package main

import (
	"errors"
	"fmt"
	"github.com/walkaba/querybuilder"
)

func main() {
	opt, err := querybuilder.FromQueryString("page[page]=0&page[size]=10&filter[companies][id]=87654321-4321-5678-8765-432187654321,null")
	if err != nil {
		fmt.Println(errors.New("invalid query string"))
	}
	queryString := querybuilder.NewQueryBuilder(true)
	filters, err := queryString.Filter(opt)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(filters)
}
