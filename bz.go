package main

import (
	"bugzilla"
)

func Bz() (*bugzilla.Conn, error) {
	return bugzilla.Dial(bugzilla.Address{
		Host:   options.Bugzilla.Host,
		Secure: options.Bugzilla.Secure,
	})
}
