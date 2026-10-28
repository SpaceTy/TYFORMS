package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	target, err := url.Parse("http://localhost:5099")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	files := http.FileServer(http.Dir("FrontEnd"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		path := filepath.Join("FrontEnd", filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			http.ServeFile(w, r, "FrontEnd/index.html")
			return
		}
		files.ServeHTTP(w, r)
	})

	addr := ":5173"
	log.Printf("Frontend dev server on http://localhost%s (proxying /api to :5099)", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
