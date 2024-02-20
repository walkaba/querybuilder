package querybuilder

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReadBuilder struct {
	collection *mongo.Collection
}

func NewSearchBuilder(collection *mongo.Collection) *ReadBuilder {
	return &ReadBuilder{collection}
}

func (c *ReadBuilder) Find(payload string) (*mongo.Cursor, error) {
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

func (c *ReadBuilder) Search(payload string) (*mongo.SingleResult, error) {
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

func (c *ReadBuilder) FindOne(id string) (*mongo.SingleResult, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", objectID}}
	result := c.collection.FindOne(context.TODO(), filter)
	return result, nil
}
