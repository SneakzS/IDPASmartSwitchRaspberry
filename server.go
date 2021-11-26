package idpa

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type wsHandler struct {
	conn *sql.DB
	mux  *sync.Mutex
}

func GetRoutes(conn *sql.DB) map[string]http.Handler {
	mux := sync.Mutex{}

	return map[string]http.Handler{
		"/ws": wsHandler{conn, &mux},
	}
}

func (s wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	p := wsProviderHandler{conn}
	err = handleProviderServer(p, s.conn)
	if err != nil {
		log.Println(err)
	}
}
