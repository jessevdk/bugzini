package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"net"
	"net/http"
	"os"
)

var router = mux.NewRouter()

var options struct {
	Debug bool `short:"d" long:"debug" description:"Enable debug mode"`

	Bugzilla struct {
		Host   string `long:"bz-host" description:"Bugzilla host (i.e. bugzilla.gnome.org)" default:"bugzilla.gnome.org"`
		Secure bool   `long:"bz-secure" description:"Use SSL" default:"true"`
	} `group:"Bugzilla Options"`
}

func SiteHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/assets/"
	router.ServeHTTP(w, r)
}

func JsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-type", "application/json")

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Could not marshal to json: %v\n", err)
	}
}

func main() {
	if _, err := flags.Parse(&options); err != nil {
		os.Exit(1)
	}

	if options.Debug {
		Assets.LocalPath = "."
	}

	router.PathPrefix("/assets/").Handler(http.FileServer(Assets))
	router.PathPrefix("/").HandlerFunc(SiteHandler)

	l, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	port := l.Addr().(*net.TCPAddr).Port

	fmt.Printf("Listening on http://localhost:%v\n", port)
	http.Serve(l, router)
}
