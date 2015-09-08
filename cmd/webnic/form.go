package main

import (
	"net"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/rafaeljusto/dnsmanager"
)

var (
	domainTemplate *template.Template
	domainRX       = regexp.MustCompile(`^((([a-z0-9][a-z0-9\-]*[a-z0-9])|[a-z0-9]+)\.)*([a-z]+|xn\-\-[a-z0-9]+)\.?$`)
)

type TemplateData struct {
	Action            string
	NewDomain         bool
	Success           bool
	Errors            map[string][]string
	RegisteredDomains []dnsmanager.Domain
	dnsmanager.Domain
}

func initializeTemplates() {
	domainTemplatePath := path.Join(config.Home, config.TemplatePath, "domain.tmpl.html")

	domainTemplate = template.Must(template.New("domain").Funcs(template.FuncMap{
		"hasErrors": func(field string, errors map[string][]string) bool {
			_, exists := errors[field]
			return exists
		},
		"getErrors": func(field string, errors map[string][]string) []string {
			return errors[field]
		},
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
		"plusplus": func(number int) int {
			number++
			return number
		},
	}).ParseFiles(domainTemplatePath))
}

func loadForm(r *http.Request) (domain dnsmanager.Domain, errs map[string][]string) {
	errs = make(map[string][]string)

	domain.FQDN = r.FormValue("domain")
	if !strings.HasSuffix(domain.FQDN, ".") {
		domain.FQDN += "."
	}

	// Limit to the top level domain
	domainParts := strings.Split(domain.FQDN, ".")
	domain.FQDN = domainParts[len(domainParts)-2] + "."

	if !domainRX.MatchString(domain.FQDN) {
		errs["domain"] = append(errs["domain"], "Invalid domain")
	}

	ns1, errsTmp := loadNS(r, domain.FQDN, "ns0")
	errs = mergeErrs(errs, errsTmp)
	domain.Nameservers = append(domain.Nameservers, ns1)

	ns2, errsTmp := loadNS(r, domain.FQDN, "ns1")
	errs = mergeErrs(errs, errsTmp)
	domain.Nameservers = append(domain.Nameservers, ns2)

	if len(ns1.Name) == 0 && len(ns2.Name) == 0 {
		errs["ns0"] = append(errs["ns0"], "At least one nameserver must be informed!")
	}

	ds1, errsTmp := loadDS(r, "ds0")
	errs = mergeErrs(errs, errsTmp)
	domain.DSSet = append(domain.DSSet, ds1)

	ds2, errsTmp := loadDS(r, "ds1")
	errs = mergeErrs(errs, errsTmp)
	domain.DSSet = append(domain.DSSet, ds2)

	return
}

func loadNS(r *http.Request, fqdn, labelPrefix string) (ns dnsmanager.Nameserver, errs map[string][]string) {
	errs = make(map[string][]string)

	if ns.Name = r.FormValue(labelPrefix); len(ns.Name) == 0 {
		return
	}

	if !strings.HasSuffix(ns.Name, ".") {
		ns.Name += "."
	}

	if !domainRX.MatchString(ns.Name) {
		errs[labelPrefix] = append(errs[labelPrefix], "Invalid nameserver")
	}

	if strings.HasSuffix(ns.Name, fqdn) {
		glue := net.ParseIP(r.FormValue(labelPrefix + "-glue"))
		if glue == nil {
			errs[labelPrefix+"-glue"] = append(errs[labelPrefix+"-glue"], "Invalid IP")
		} else {
			ns.Glues = append(ns.Glues, glue)
		}
	} else {
		errs[labelPrefix] = append(errs[labelPrefix], "For our lab please use a name that needs a glue")
	}

	return
}

func loadDS(r *http.Request, labelPrefix string) (ds dnsmanager.DS, errs map[string][]string) {
	errs = make(map[string][]string)

	if len(r.FormValue(labelPrefix+"-digest")) == 0 {
		return
	}

	if keyTag, err := strconv.ParseUint(r.FormValue(labelPrefix+"-keytag"), 10, 16); err == nil {
		ds.KeyTag = uint16(keyTag)
	} else {
		errs[labelPrefix+"-keytag"] = append(errs[labelPrefix+"-keytag"], "Must be a number!")
	}

	if algorithm, err := strconv.ParseUint(r.FormValue(labelPrefix+"-algorithm"), 10, 8); err == nil {
		ds.Algorithm = uint8(algorithm)
	} else {
		errs[labelPrefix+"-algorithm"] = append(errs[labelPrefix+"-algorithm"], "Must be a number!")
	}

	if digestType, err := strconv.ParseUint(r.FormValue(labelPrefix+"-digest-type"), 10, 8); err == nil {
		ds.DigestType = uint8(digestType)
	} else {
		errs[labelPrefix+"-digest-type"] = append(errs[labelPrefix+"-digest-type"], "Must be a number!")
	}

	ds.Digest = strings.Replace(strings.ToUpper(r.FormValue(labelPrefix+"-digest")), " ", "", -1)

	return
}

func mergeErrs(errs1, errs2 map[string][]string) map[string][]string {
	for fqdn, list := range errs2 {
		for _, item := range list {
			errs1[fqdn] = append(errs1[fqdn], item)
		}
	}
	return errs1
}
