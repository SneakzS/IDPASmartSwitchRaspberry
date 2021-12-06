package serverui

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/julienschmidt/httprouter"
)

var (
	//go:embed templates/*
	templateFiles embed.FS
	//go:embed static/*
	static embed.FS
)

type dummyfs struct{}

var _ fs.FS = dummyfs{}

func (dummyfs) Open(name string) (fs.File, error) {
	fmt.Println(name)
	return nil, os.ErrNotExist
}

func GetHTTPHandler(conn *sql.DB) http.Handler {
	r := httprouter.New()

	getIndexRoutes(r)
	getCustomerRoutes(r, conn)

	r.Handler("GET", "/static/*filepath", http.FileServer(http.FS(static)))

	//http.Handle(staticPath)

	return r
}

type baseView struct {
	Title string
}

func (baseView) Static(resource string) string {
	return path.Join("/static/", resource)
}

func compileTemplate(files ...string) *template.Template {
	patterns := make([]string, len(files))
	for i, file := range files {
		patterns[i] = "templates/" + file
	}
	tpl := template.Must(template.ParseFS(templateFiles, patterns...))
	return tpl
}
