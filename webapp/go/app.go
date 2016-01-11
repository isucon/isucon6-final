package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var staticPath string

func init() {
	flag.StringVar(&staticPath, "staticpath", "", "static file directory")
}

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func main() {
	flag.Parse()

	goji.Get("/hello/:name", hello)
	goji.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	goji.Serve()
}
