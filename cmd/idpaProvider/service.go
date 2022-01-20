package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"path"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/philip-s/idpa/common"
	"github.com/philip-s/idpa/provider"
)

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

func runService() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}

	c := provider.Config{}
	err = provider.ReadConfig(&c, ini)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", c.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	serverMutex := sync.Mutex{}

	r := httprouter.New()
	r.Handler("GET", "/static/*filepath", http.FileServer(http.FS(static)))
	r.HandlerFunc("POST", "/api/v1/workload", func(rw http.ResponseWriter, r *http.Request) {
		provider.WorkloadRequestHandler(rw, r, db, &serverMutex)
	})
	addCustomerRoutes(r, db)
	addIndexRoutes(r)

	return http.ListenAndServe(c.Listen, r)
}
