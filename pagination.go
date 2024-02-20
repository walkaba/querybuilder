package querybuilder

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	result.Data = cursor
	var meta Meta
	meta.Page.CurrentPage = page
	meta.Page.PerPage = size
	meta.Page.Total = count
	meta.Filters = payload
	bytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	result.Meta = bytes
	return &result, nil
}
