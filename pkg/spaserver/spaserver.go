package spaserver

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type Routes map[string]http.HandlerFunc

type SPAServer struct {
	r  *Routes
	mu *http.ServeMux
}

func NewSPAServer(r *Routes, routePrefix, fileServerPath string) *SPAServer {
	mu := http.NewServeMux()
	routes(mu, r, routePrefix, fileServerPath)
	return &SPAServer{r, mu}
}

func (s *SPAServer) Listen(port string) {
	http.ListenAndServe(port, s.mu)
}

func routes(mu *http.ServeMux, r *Routes, px, sp string) {
	for p, handler := range *r {
		route(mu, px, p, handler)
	}
	initiateSPAServer(mu, sp)
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

func initiateSPAServer(mu *http.ServeMux, pp string) {
	pp = path.Clean(pp)
	ap := path.Join(pp, "/assets")
	spah := http.Handler(spaHandler{pp})
	mu.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(ap))))
	mu.Handle("/", spah)
	log.Printf("routing assets FileServer of path \"%v\" on /assets\n", ap)
	log.Println("routing SPA handler on /")
	log.Println(os.ReadDir(ap))
}

type spaHandler struct {
	pp string
}

func (spah spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	index, err := os.ReadFile(path.Join(spah.pp, "/index.html"))

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("can`t find an index"))
		return
	}

	w.WriteHeader(200)
	// w.Header().Add("Content-Type", "text/html")
	w.Write(index)

}
