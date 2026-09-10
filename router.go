package main

import (
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const maxCommentLen = 50

// allowedOrigins 是允许跨域携带凭证的来源白名单，由 ALLOWED_ORIGINS 配置
// （逗号分隔，需写完整来源，例如 https://chat.example.com）。
//
// 认证改用 HttpOnly Cookie 后，响应里不能再回 Access-Control-Allow-Origin: *，
// 因为浏览器不允许 "*" 与凭证同时使用。所以这里改为回显白名单内的具体来源。
// 未配置时不下发任何 CORS 头，即只允许同源访问。
var allowedOrigins = parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))

func parseAllowedOrigins(raw string) map[string]bool {
	origins := make(map[string]bool)
	for _, part := range strings.Split(raw, ",") {
		if origin := strings.TrimSpace(part); origin != "" {
			origins[origin] = true
		}
	}
	return origins
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin != "" && allowedOrigins[origin]

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			// 响应随 Origin 变化，必须声明，否则中间缓存会把 A 站的响应给 B 站。
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			if allowed {
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
				c.Header("Access-Control-Max-Age", "600")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			// 不在白名单内的预检直接拒绝，且不带任何 CORS 头。
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func indexRouter(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func commentsGetRouter(c *gin.Context) {
	c.JSON(http.StatusOK, getComments())
}

// commentsPostRouter 需要登录，用户名一律取自 JWT，忽略请求体里的 name。
func commentsPostRouter(c *gin.Context) {
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
