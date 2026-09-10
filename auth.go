package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	// tokenTTL 是签发出去的 JWT 有效期。
	tokenTTL = 7 * 24 * time.Hour
	// tokenIssuer 用于标识 token 的签发方。
	tokenIssuer = "freechat"

	minUsernameLen = 3
	maxUsernameLen = 20
	minPasswordLen = 6
	maxPasswordLen = 72 // bcrypt 只使用前 72 字节，超长直接拒绝以免静默截断

	ctxUsernameKey = "authUsername"
	ctxUserIDKey   = "authUserID"
)

// fallbackJWTSecret 只用于本地开发。生产环境必须通过 JWT_SECRET 覆盖。
var fallbackJWTSecret = []byte("freechat-insecure-dev-secret-please-override")

var (
	jwtSecret          []byte
	jwtSecretFromEnv   bool
	errInvalidToken    = errors.New("invalid token")
	errUnexpectedAlg   = errors.New("unexpected signing method")
	errInvalidUsername = errors.New("用户名长度需为 3-20 个字符")
	errInvalidPassword = errors.New("密码长度需为 6-72 个字符")
)

func init() {
	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		jwtSecret = []byte(secret)
		jwtSecretFromEnv = true
		return
	}
	jwtSecret = fallbackJWTSecret
}

// claims 是写进 JWT 的负载，只放身份信息，不放任何敏感数据。
type claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authPayload struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func generateToken(user User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(tokenTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(jwtSecret)
	return signed, expiresAt, err
}

func parseToken(tokenString string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedAlg
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	parsed, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, errInvalidToken
	}
	if parsed.Username == "" || parsed.UserID == 0 {
		return nil, errInvalidToken
	}
	return parsed, nil
}

// bearerToken 从 Authorization 头里取出 Bearer token。
func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// authRequired 校验 JWT，通过后把用户信息写入 gin.Context。
func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "请先登录"})
			return
		}

		parsed, err := parseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "登录状态已失效，请重新登录"})
			return
		}

		c.Set(ctxUsernameKey, parsed.Username)
		c.Set(ctxUserIDKey, parsed.UserID)
		c.Next()
	}
}

func currentUsername(c *gin.Context) string {
	if value, ok := c.Get(ctxUsernameKey); ok {
		if username, ok := value.(string); ok {
			return username
		}
	}
	return ""
}

func validateCredentials(cred credentials) error {
	length := utf8.RuneCountInString(cred.Username)
	if length < minUsernameLen || length > maxUsernameLen {
		return errInvalidUsername
	}
	if utf8.RuneCountInString(cred.Password) < minPasswordLen || len(cred.Password) > maxPasswordLen {
		return errInvalidPassword
	}
	return nil
}

func registerRouter(c *gin.Context) {
	var cred credentials
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式有误"})
		return
	}

	cred.Username = strings.TrimSpace(cred.Username)
	if err := validateCredentials(cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if _, exists := getUserByName(cred.Username); exists {
		c.JSON(http.StatusConflict, gin.H{"message": "用户名已被占用"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "注册失败，请稍后重试"})
		return
	}

	user, err := createUser(cred.Username, string(hash))
	if err != nil {
		// unique_index 冲突：并发注册同名用户时兜底。
		c.JSON(http.StatusConflict, gin.H{"message": "用户名已被占用"})
		return
	}

	token, expiresAt, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "注册失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusCreated, authPayload{Token: token, Username: user.Username, ExpiresAt: expiresAt})
}

func loginRouter(c *gin.Context) {
	var cred credentials
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式有误"})
		return
	}

	cred.Username = strings.TrimSpace(cred.Username)
	user, exists := getUserByName(cred.Username)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cred.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, expiresAt, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "登录失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, authPayload{Token: token, Username: user.Username, ExpiresAt: expiresAt})
}

// meRouter 让前端在刷新页面后校验本地 token 是否仍然有效。
func meRouter(c *gin.Context) {
	id, _ := c.Get(ctxUserIDKey)
	userID, _ := id.(uint)
	c.JSON(http.StatusOK, gin.H{"username": currentUsername(c), "id": userID})
}
