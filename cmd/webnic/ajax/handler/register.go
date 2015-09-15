package handler

import "github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy"

var (
	Routes map[string]handy.Constructor
)

func handle(pattern string, handler handy.Constructor) {
	if Routes == nil {
		Routes = make(map[string]handy.Constructor)
	}

	Routes[pattern] = handler
}
