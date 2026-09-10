package main

import (
	"strings"

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

// User 是注册用户，密码只保存 bcrypt 摘要，绝不明文落库。
type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(32);unique_index" json:"username"`
	PasswordHash string `gorm:"type:varchar(100)" json:"-"`
}

func getComments() []Comment {
	var comments []Comment
	db.Find(&comments)
	return comments
}

func addNewComment(name string, content string, os string) {
	db.Create(&Comment{Name: name, Content: content, OS: os})
}

// getUserByName 按用户名查找用户。比较时忽略大小写，否则 "Admin" 与 "admin"
// 会变成两个账号，形成冒名注册的漏洞。
// 注意：SQLite 的 LOWER() 只折叠 ASCII 大小写，非 ASCII 用户名按原文比较。
func getUserByName(username string) (User, bool) {
	var user User
	err := db.Where("LOWER(username) = ?", strings.ToLower(username)).First(&user).Error
	return user, err == nil
}

func createUser(username string, passwordHash string) (User, error) {
	user := User{Username: username, PasswordHash: passwordHash}
	err := db.Create(&user).Error
	return user, err
}

// ensureUserIndexes 补上 AutoMigrate 无法表达的表达式索引。
// users.username 上的唯一索引是大小写敏感的，这里再加一把按 LOWER(username)
// 的唯一索引，把“忽略大小写判重”从 Go 层的检查变成数据库层的硬约束，
// 从而堵住并发注册 "Admin" / "admin" 的竞态。
func ensureUserIndexes() error {
	return db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS uix_users_username_lower ON users (LOWER(username))",
	).Error
}
