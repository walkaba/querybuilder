package querybuilder

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"strings"
)

type PaginationBuilder struct {
	collection *mongo.Collection
	route      string
}

func NewPaginationBuilder(collection *mongo.Collection, route string) *PaginationBuilder {
	return &PaginationBuilder{collection, route}
}

func (c *PaginationBuilder) Find(payload string) (*mongo.Cursor, error) {
	opt, err := FromQueryString(payload)
	if err != nil {
		return nil, errors.New("invalid query string")
	}
	queryString := NewQueryBuilder(true)
	options, err := queryString.FindOptions(opt)
	if err != nil {
		return nil, err
	}
	var filters bson.D
	if len(opt.Filter) > 0 {
		filters, err = queryString.Filter(opt)
		if err != nil {
			return nil, err
		}
	} else {
		filters = bson.D{}
	}
	cursor, err := c.collection.Find(context.TODO(), filters, options)
	if err != nil {
		if err.Error() != "document is nil" {
			return nil, err
		}
	}
	return cursor, nil
}

func (c *PaginationBuilder) FindOne(payload string) (*mongo.SingleResult, error) {
	opt, err := FromQueryString(payload)
	if err != nil {
		return nil, errors.New("invalid query string")
	}
	queryString := NewQueryBuilder(true)
	var filters bson.D
	if len(opt.Filter) > 0 {
		filters, err = queryString.Filter(opt)
		if err != nil {
			return nil, err
		}
	} else {
		filters = bson.D{}
	}
	result := c.collection.FindOne(context.TODO(), filters)
	return result, nil
}

func (c *PaginationBuilder) Pagination(payload string) (*OutPagination, error) {
	opt, err := FromQueryString(payload)
	if err != nil {
		return nil, errors.New("invalid query string")
	}
	queryString := NewQueryBuilder(true)
	findOptions, err := queryString.FindOptions(opt)
	if err != nil {
		return nil, err
	}
	var filters bson.D
	if len(opt.Filter) > 0 {
		filters, err = queryString.Filter(opt)
		if err != nil {
			return nil, err
		}
	} else {
		filters = bson.D{}
	}
	count, err := c.collection.CountDocuments(context.TODO(), filters)
	if err != nil {
		if err.Error() != "document is nil" {
			return nil, err
		}
	}
	cursor, err := c.collection.Find(context.TODO(), filters, findOptions)
	if err != nil {
		if err.Error() != "document is nil" {
			return nil, err
		}
	}
	var result OutPagination
	page := int64(opt.Page["page"])
	size := int64(opt.Page["size"])
	skip := size * (page - 1)
	if (page + 1) <= count {
		result.Meta.Page.LastPage = page + 1
	} else {
		result.Meta.Page.LastPage = page
	}
	result.Data = cursor
	result.Meta.Page.CurrentPage = page
	result.Meta.Page.PerPage = size
	result.Meta.Page.Total = count
	result.Meta.Page.From = skip + 1
	result.Meta.Page.To = skip + size
	result.Meta.Links.First = c.generateLink(0, size) + c.generateFilters(opt.Filter)
	result.Meta.Links.Last = c.generateLink(result.Meta.Page.LastPage, size) + c.generateFilters(opt.Filter)
	result.Meta.Links.Prev = c.generateLink(result.Meta.Page.CurrentPage-1, size) + c.generateFilters(opt.Filter)
	result.Meta.Links.Next = c.generateLink(result.Meta.Page.CurrentPage+1, size) + c.generateFilters(opt.Filter)
	result.Meta.Filters = payload
	return &result, nil
}

func (c *PaginationBuilder) generateLink(page, size int64) string {
	return "/" + c.route + "/?page[page]=" + strconv.FormatInt(page, 10) + "&page[size]=" + strconv.FormatInt(size, 10)
}

func (c *PaginationBuilder) generateFilters(filters map[string][]string) string {
	var result string
	for key, value := range filters {
		isArray := strings.Split(key, "][")
		if len(isArray) > 1 {
			result += "&filter[" + isArray[0] + "][" + isArray[1] + "]=" + strings.Join(value, ",")
		} else {
			result += "&filter[" + key + "]=" + strings.Join(value, ",")
		}
	}
	return result
}
