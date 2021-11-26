package serverui

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"path"
)

var (
	//go:embed templates/*
	templateFiles embed.FS
	//go:embed static/*
	static embed.FS

	staticPath = "/static/"
)

func GetRoutes(conn *sql.DB) map[string]http.Handler {
	return map[string]http.Handler{
		staticPath:   http.FileServer(http.FS(static)),
		"/":          indexController{conn},
		"/customers": customerController{conn},
	}
}

type baseView struct {
	Title string
}

func (baseView) Static(resource string) string {
	return path.Join(staticPath, resource)
}

func compileTemplate(files ...string) *template.Template {
	patterns := make([]string, len(files))
	for i, file := range files {
		patterns[i] = "templates/" + file
	}
	tpl := template.Must(template.ParseFS(templateFiles, patterns...))
	return tpl
}
