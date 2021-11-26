package serverui

import (
	"database/sql"
	"log"
	"net/http"
)

type indexController struct {
	conn *sql.DB
}

type indexView struct {
	baseView
	Message string
}

var indexTemplate = compileTemplate("layout.html", "index.html")

func (c indexController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view := indexView{Message: "Hello World"}
	view.Title = "Home"
	err := indexTemplate.Execute(w, view)
	if err != nil {
		log.Println(err)
	}
}
