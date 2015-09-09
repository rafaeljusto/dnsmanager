package handler

import (
	"html/template"
	"strconv"

	"github.com/rafaeljusto/dnsmanager"
)

type templateData struct {
	dnsmanager.Domain

	Action            string
	RegisteredDomains []dnsmanager.Domain
	NewDomain         bool
	Success           bool
	Errors            map[string][]string
}

func NewTemplateData() templateData {
	return templateData{
		Errors: make(map[string][]string),
	}
}

var templateFuncs = template.FuncMap{
	"getNameserver": func(index int) string {
		return "ns" + strconv.Itoa(index)
	},
	"getGlue": func(index int) string {
		return "ns" + strconv.Itoa(index) + "-glue"
	},
	"getKeytag": func(index int) string {
		return "ds" + strconv.Itoa(index) + "-keytag"
	},
	"getAlgorithm": func(index int) string {
		return "ds" + strconv.Itoa(index) + "-algorithm"
	},
	"getDigestType": func(index int) string {
		return "ds" + strconv.Itoa(index) + "-digest-type"
	},
	"getDigest": func(index int) string {
		return "ds" + strconv.Itoa(index) + "-digest"
	},
	"print": func(domain *dnsmanager.Domain) string {
		nsCount := 0
		for _, nameserver := range domain.Nameservers {
			if len(nameserver.Name) > 0 {
				nsCount++
			}
		}

		dsCount := 0
		for _, ds := range domain.DSSet {
			if len(ds.Digest) > 0 {
				dsCount++
			}
		}

		return strconv.Itoa(nsCount) + " NS, " + strconv.Itoa(dsCount) + " DS"
	},
}
