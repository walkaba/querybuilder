package querybuilder

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
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
	filter := bson.D{{"_id", objectID}}
	_, err = c.collection.UpdateOne(context.TODO(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	return err
}

func (c *WriteBuilder) UpdateOne(id string, body interface{}) (*string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", objectID}}
	_, err = c.collection.UpdateOne(context.TODO(), filter, body)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (c *WriteBuilder) InsertOne(body interface{}) (*string, error) {
	now := time.Now()
	document := bson.D{
		{"$set", body},
		{"createdAt", now},
		{"updatedAt", now},
	}
	result, err := c.collection.InsertOne(context.TODO(), document)
	if err != nil {
		return nil, err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return &id, nil
}
