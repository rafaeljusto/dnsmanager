package handler

import "github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy"

type defaultHandler struct {
	uriVars handy.URIVars
}

func (d *defaultHandler) URIVars() handy.URIVars {
	return d.uriVars
}
