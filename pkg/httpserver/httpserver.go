package httpserver

import (
	"log"
	"net/http"
	"strings"
)

type Routes map[string]http.HandlerFunc

type Server struct {
	r  *Routes
	mu *http.ServeMux
}

func NewServer(r *Routes, routePrefix, fileServerPath string) *Server {
	mu := http.NewServeMux()
	routes(mu, r, routePrefix, fileServerPath)
	return &Server{r, mu}
}

func (s *Server) Listen(port string) {
	// s.mu.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(200)
	// 	w.Header().Add("Content-Type", "application/json")
	// 	w.Write([]byte("{\"ok\": true}"))
	// })

	http.ListenAndServe(port, s.mu)
}

func routes(mu *http.ServeMux, r *Routes, px, sp string) {
	for p, handler := range *r {
		route(mu, px, p, handler)
	}
	initiateFileServer(mu, sp)
}

func route(mu *http.ServeMux, px string, p string, handler http.HandlerFunc) {
	var method, path string
	path = px + p
	if mp := strings.Split(p, " "); len(mp) > 1 {
		method = mp[0] + " "
		path = px + mp[1]
	}

	log.Printf("routing path \"%v\"\n", method+path)
	mu.Handle(method+path, handler)
}

func initiateFileServer(mu *http.ServeMux, p string) {
	log.Printf("routing file server of path \"%v\" to route /\n", p)

	mu.Handle("GET /", http.FileServer(http.Dir(p)))
}
