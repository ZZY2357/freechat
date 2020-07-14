package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func indexRouter(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func commentsGetRouter(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, getComments())
}

func commentsPostRouter(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var comment Comment
	c.BindJSON(&comment)
	addNewComment(comment.Name, comment.Content)
	c.JSON(http.StatusOK, getComments())
}
