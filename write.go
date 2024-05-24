package querybuilder

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WriteBuilder struct {
	collection *mongo.Collection
}

func NewWriteBuilder(collection *mongo.Collection) *WriteBuilder {
	return &WriteBuilder{collection}
}

func (c *WriteBuilder) DeleteOne(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	_, err = c.collection.UpdateOne(context.TODO(), filter, bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now()}}}})
	return err
}

func (c *WriteBuilder) UpdateOne(id string, update bson.M) (*string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	now := time.Now()
	if setFields, ok := update["$set"].(bson.M); ok {
		setFields["updatedAt"] = now
	}
	_, err = c.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (c *WriteBuilder) InsertOne(body interface{}) (*string, error) {
	now := time.Now()
	bodyMap := bson.M{}
	bodyBytes, err := bson.Marshal(body)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		return nil, err
	}
	bodyMap["createdAt"] = now
	bodyMap["updatedAt"] = now
	result, err := c.collection.InsertOne(context.TODO(), bodyMap)
	if err != nil {
		return nil, err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return &id, nil
}
