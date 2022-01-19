package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type indexView struct {
	baseView
	Message string
}

var indexTemplate = compileTemplate("layout.html", "index.html")

func addIndexRoutes(r *httprouter.Router) {
	r.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		view := indexView{Message: "Hello World"}
		view.Title = "Home"
		err := indexTemplate.Execute(w, view)
		if err != nil {
			log.Println(err)
		}
	})
}
