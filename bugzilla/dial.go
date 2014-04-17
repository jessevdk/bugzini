package bugzilla

import (
	"fmt"
	"github.com/jessevdk/xmlrpc"
)

type Conn struct {
	*xmlrpc.Client
}

type Address struct {
	Host   string
	Port   int
	Secure bool
}

func (a *Address) uri() string {
	var ret string

	if a.Secure {
		ret = "https://"
	} else {
		ret = "http://"
	}

	ret += a.Host

	if a.Port > 0 {
		ret += fmt.Sprintf("%v", a.Port)
	}

	ret += "/xmlrpc.cgi"
	return ret
}

func Dial(address Address) (*Conn, error) {
	client, err := xmlrpc.NewClient(address.uri(), nil)

	if err != nil {
		return nil, err
	}

	return &Conn{
		Client: client,
	}, nil
}
