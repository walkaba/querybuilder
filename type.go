package querybuilder

import "go.mongodb.org/mongo-driver/mongo"

type Page struct {
	CurrentPage int64 `json:"currentPage"`
	PerPage     int64 `json:"perPage"`
	From        int64 `json:"from"`
	To          int64 `json:"to"`
	Total       int64 `json:"total"`
	LastPage    int64 `json:"lastPage"`
}

type Link struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

type Meta struct {
	Page    Page        `json:"page"`
	Links   Link        `json:"links"`
	Filters interface{} `json:"filters"`
}

type OutPagination struct {
	Data *mongo.Cursor `json:"data"`
	Meta Meta          `json:"meta"`
}
