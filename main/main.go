package main

import (
	"context"
	"fmt"
	"github.com/walkaba/querybuilder"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func main() {
	conn, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://walkaba:A1fUnY7yPtaccmWJ@walkaba.sacs6nv.mongodb.net/?retryWrites=true&w=majority"))
	collection := conn.Database(os.Getenv("MONGO_DATABASE")).Collection("profiles")
	col := querybuilder.NewPaginationBuilder(collection, "profiles")
	result, _ := col.Pagination("page[page]=0&page[size]=10&filter[companies][id]=87654321-4321-5678-8765-432187654321,null")
	fmt.Println(result)

	opt, _ := querybuilder.FromQueryString("page[page]=0&page[size]=10&filter[companies][id]=87654321-4321-5678-8765-432187654321,null")
	queryString := querybuilder.NewQueryBuilder(true)
	filters, _ := queryString.Filter(opt)
	fmt.Println(filters)
}
