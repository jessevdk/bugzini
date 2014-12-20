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

const BugziniVersion = uint32(1)

var router = mux.NewRouter()

var options struct {
	Debug  bool `short:"d" long:"debug" description:"Enable debug mode"`
	Launch bool `short:"l" long:"launch" description:"Launch browser at location"`
	Port   int  `short:"p" long:"port" description:"Launch local webserver at specified port" default:"8000"`

	Bugzilla struct {
		Host   string `long:"bz-host" description:"Bugzilla host (i.e. bugzilla.gnome.org)" default:"bugzilla.gnome.org"`
		Secure bool   `long:"bz-secure" description:"Use SSL" default:"true"`
	} `group:"Bugzilla Options"`
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies
}

func SiteHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/assets/"
	router.ServeHTTP(w, r)
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	noCache(w)

	info := struct {
		Version uint32 `json:"version"`
		Host    string `json:"host"`
		Secure  bool   `json:"secure"`
	}{
		Version: BugziniVersion,
		Host:    options.Bugzilla.Host,
		Secure:  options.Bugzilla.Secure,
	}

	JsonResponse(w, info)
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

	cache.Load()

	if options.Debug {
		Assets.LocalPath = "."
	}

	var bzaddr string

	if options.Bugzilla.Secure {
		bzaddr = "https://"
	} else {
		bzaddr = "http://"
	}

	bzaddr += options.Bugzilla.Host

	router.Handle("/favicon.ico", http.RedirectHandler(bzaddr+"/favicon.ico", http.StatusTemporaryRedirect))
	router.HandleFunc("/api/info", InfoHandler)

	router.PathPrefix("/assets/").Handler(http.FileServer(Assets))
	router.PathPrefix("/").HandlerFunc(SiteHandler)

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", options.Port))

	if err != nil {
		panic(err)
	}

	port := l.Addr().(*net.TCPAddr).Port

	fmt.Printf("Listening on http://localhost:%v\n", port)

	if options.Launch {
		launchBrowser(port)
	}

	http.Serve(l, router)
}
