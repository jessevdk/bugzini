package main

import (
	"bugzilla"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/jessevdk/xmlrpc"
)

var bz *bugzilla.Conn

func SaveCookies() {
	if bz == nil || bz.Client == nil {
		return
	}

	cookies := bz.Client.Cookies()
	host := bz.Client.CookieHost()

	os.MkdirAll(".cookies", 0700)

	f, err := os.OpenFile(path.Join(".cookies", host), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cookies: %s\n", err)
		return
	}

	defer f.Close()

	enc := gob.NewEncoder(f)

	if err := enc.Encode(cookies); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode cookies: %s\n", err)
	}
}

func loadSavedCookies(client *xmlrpc.Client) {
	host := client.CookieHost()

	f, err := os.Open(path.Join(".cookies", host))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open cookie file: %s\n", err)
		return
	}

	defer f.Close()
	dec := gob.NewDecoder(f)

	cookies := make([]*http.Cookie, 0, 10)

	if err := dec.Decode(&cookies); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode cookies: %s\n", err)
		return
	}

	client.SetCookies(cookies)
}

func Bz() (*bugzilla.Conn, error) {
	if bz != nil {
		return bz, nil
	}

	var err error

	bz, err = bugzilla.Dial(bugzilla.Address{
		Host:   options.Bugzilla.Host,
		Secure: options.Bugzilla.Secure,
	})

	loadSavedCookies(bz.Client)

	return bz, err
}
