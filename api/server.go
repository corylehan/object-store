// api/server.go
package api

import (
    "fmt"
    "net/http"

    "github.com/corylehan/object-store/store"
)

type Server struct {
    Port   int
    Store  *store.Store
    Router *http.ServeMux
}

func NewServer(port int, s *store.Store) *Server {
    server := &Server{
        Port:   port,
        Store:  s,
        Router: http.NewServeMux(),
    }

    h := NewHandler(s)
    server.Router.HandleFunc("/objects", h.handleObjects)
    server.Router.HandleFunc("/objects/", h.handleObject)

    return server
}

func (s *Server) ListenAndServe() error {
    return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router)
}
