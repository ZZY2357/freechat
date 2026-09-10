package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 前端资源直接编进二进制，这样把 freechat.exe 单独拷到任何目录都能跑，
// 不再依赖运行目录下的 views/ 和 static/。
//
// 代价：改前端文件后必须重新 go build 才生效。
//
//go:embed views static
var embeddedAssets embed.FS

// htmlDelims 是模板分隔符。用 <!-- --> 是为了和 Vue 的 {{ }} 插值共存。
func htmlDelims() (string, string) {
	return "<!--", "-->"
}

// registerAssets 把内嵌的模板和静态资源挂到 gin 引擎上。
func registerAssets(r *gin.Engine) error {
	left, right := htmlDelims()

	// ParseFS 用 filepath.Base 作为模板名，所以 views/index.html 注册为 "index.html"
	tmpl, err := template.New("").Delims(left, right).ParseFS(embeddedAssets, "views/*")
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(tmpl)

	staticFS, err := fs.Sub(embeddedAssets, "static")
	if err != nil {
		return err
	}
	r.StaticFS("/static", http.FS(staticFS))

	return nil
}
