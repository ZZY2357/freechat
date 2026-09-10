package main

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const maxCommentLen = 50

func indexRouter(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func commentsGetRouter(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, getComments())
}

// commentsPostRouter 需要登录，用户名一律取自 JWT，忽略请求体里的 name。
func commentsPostRouter(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var comment Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式有误"})
		return
	}

	content := strings.TrimSpace(comment.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "内容不能为空"})
		return
	}
	if utf8.RuneCountInString(content) > maxCommentLen {
		c.JSON(http.StatusBadRequest, gin.H{"message": "内容不能超过 50 个字符"})
		return
	}

	addNewComment(currentUsername(c), content, comment.OS)
	c.JSON(http.StatusOK, getComments())
}
