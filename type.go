package querybuilder

import "go.mongodb.org/mongo-driver/mongo"

type Page struct {
	CurrentPage int64 `json:"currentPage"`
	PerPage     int64 `json:"perPage"`
	Total       int64 `json:"total"`
}

type Meta struct {
	Page    Page        `json:"page"`
	Filters interface{} `json:"filters"`
}

type OutPagination struct {
	Data *mongo.Cursor `json:"data"`
	Meta []byte        `json:"meta"`
}
