package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db, err = gorm.Open("sqlite3", "./data.db")

type Comment struct {
	gorm.Model
	Name    string `json:"name"`
	Content string `json:"content"`
	OS      string `json:"os"`
}

func getComments() []Comment {
	var comments []Comment
	db.Find(&comments)
	return comments
}

func addNewComment(name string, content string, os string) {
	db.Create(&Comment{Name: name, Content: content, OS: os})
}
