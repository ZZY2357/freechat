package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	defer db.Close()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Comment{}, &User{})
	if indexErr := ensureUserIndexes(); indexErr != nil {
		log.Printf("[warn] 创建用户名唯一索引失败，可能存在仅大小写不同的重名账号: %v", indexErr)
	}

	if !jwtSecretFromEnv {
		log.Println("[warn] 未设置 JWT_SECRET 环境变量，当前使用内置默认密钥，请勿用于生产环境")
	}

	r := gin.Default()
	r.Static("static", "static")
	r.Delims("<!--", "-->")
	r.LoadHTMLGlob("views/*")

	r.GET("/", indexRouter)
	r.GET("/comments", commentsGetRouter)
	r.POST("/comments", authRequired(), commentsPostRouter)

	r.POST("/auth/register", registerRouter)
	r.POST("/auth/login", loginRouter)
	r.GET("/auth/me", authRequired(), meRouter)

	r.Run(":8080")
}
