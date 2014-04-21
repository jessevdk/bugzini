package main

import (
	"bugzilla"
)

var bz *bugzilla.Conn

func Bz() (*bugzilla.Conn, error) {
	if bz != nil {
		return bz, nil
	}

	var err error

	bz, err = bugzilla.Dial(bugzilla.Address{
		Host:   options.Bugzilla.Host,
		Secure: options.Bugzilla.Secure,
	})

	return bz, err
}
