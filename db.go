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
}

func getComments() []Comment {
	var users []Comment
	db.Find(&users)
	return users
}

func addNewComment(name string, content string) {
	db.Create(&Comment{Name: name, Content: content})
}
