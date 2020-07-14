package main

import "github.com/gin-gonic/gin"

func main() {
	defer db.Close()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Comment{})

	r := gin.New()
	r.Static("static", "static")
	r.Delims("<!--", "-->")
	r.LoadHTMLGlob("views/*")

	r.GET("/", indexRouter)
	r.GET("/comments", commentsGetRouter)
	r.POST("/comments", commentsPostRouter)

	r.Run(":8080")
}
