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
	if !cookieSecure {
		log.Println("[warn] COOKIE_SECURE 未开启，认证 Cookie 会随明文 HTTP 传输；上 HTTPS 后请设为 true")
	}
	if len(allowedOrigins) == 0 {
		log.Println("[info] 未配置 ALLOWED_ORIGINS，仅允许同源访问跨域接口")
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	if err := registerAssets(r); err != nil {
		log.Fatalf("加载内嵌前端资源失败: %v", err)
	}

	r.GET("/", indexRouter)
	r.GET("/comments", commentsGetRouter)
	r.POST("/comments", authRequired(), commentsPostRouter)

	r.POST("/auth/register", registerRouter)
	r.POST("/auth/login", loginRouter)
	r.POST("/auth/logout", logoutRouter)
	r.GET("/auth/me", authRequired(), meRouter)

	r.Run(":8080")
}
