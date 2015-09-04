package main

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var (
	domainTemplate *template.Template
	isDomain       = regexp.MustCompile("([a-zA-Z0-9]\\.)*([a-zA-Z0-9](\\.)?)")
	isIP           = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}")

	errNotFound = errors.New("No records for domain")
)

type TemplateData struct {
	Action     string
	NewDomain  bool
	Success    bool
	Errors     map[string][]string
	Subdomains map[string]*Domain
	Domain
}

func init() {
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
		"print": func(domain *Domain) string {
			nsCount := 0
			for _, nameserver := range domain.Nameservers {
				if len(nameserver.Name) > 0 {
					nsCount++
				}
			}

			dsCount := 0
			for _, ds := range domain.DSs {
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
	}).ParseFiles("templates/domain.tmpl.html"))
}

func loadForm(r *http.Request) (domain *Domain, errs map[string][]string) {
	domain = new(Domain)
	errs = make(map[string][]string)

	domain.Name = r.FormValue("domain")
	if !strings.HasSuffix(domain.Name, ".") {
		domain.Name += "."
	}

	// Limit to the top level domain
	domainParts := strings.Split(domain.Name, ".")
	domain.Name = domainParts[len(domainParts)-2] + "."

	if !isDomain.MatchString(domain.Name) {
		errs["domain"] = append(errs["domain"], "Invalid domain")
	}

	ns1, errsTmp := loadNS(r, domain.Name, "ns0")
	errs = mergeErrs(errs, errsTmp)
	domain.Nameservers = append(domain.Nameservers, ns1)

	ns2, errsTmp := loadNS(r, domain.Name, "ns1")
	errs = mergeErrs(errs, errsTmp)
	domain.Nameservers = append(domain.Nameservers, ns2)

	if len(ns1.Name) == 0 && len(ns2.Name) == 0 {
		errs["ns0"] = append(errs["ns0"], "At least one nameserver must be informed!")
	}

	ds1, errsTmp := loadDS(r, "ds0")
	errs = mergeErrs(errs, errsTmp)
	domain.DSs = append(domain.DSs, ds1)

	ds2, errsTmp := loadDS(r, "ds1")
	errs = mergeErrs(errs, errsTmp)
	domain.DSs = append(domain.DSs, ds2)

	return
}

func loadNS(r *http.Request, fqdn, labelPrefix string) (ns *Nameserver, errs map[string][]string) {
	ns = new(Nameserver)
	errs = make(map[string][]string)

	if ns.Name = r.FormValue(labelPrefix); len(ns.Name) == 0 {
		return
	}

	if !strings.HasSuffix(ns.Name, ".") {
		ns.Name += "."
	}

	if !isDomain.MatchString(ns.Name) {
		errs[labelPrefix] = append(errs[labelPrefix], "Invalid nameserver")
	}

	if strings.HasSuffix(ns.Name, fqdn) {
		ns.Glue = r.FormValue(labelPrefix + "-glue")
		if !isIP.MatchString(ns.Glue) {
			errs[labelPrefix+"-glue"] = append(errs[labelPrefix+"-glue"], "Invalid IP")
		}
	} else {
		errs[labelPrefix] = append(errs[labelPrefix], "For our lab please use a name that needs a glue")
	}

	return
}

func loadDS(r *http.Request, labelPrefix string) (ds *DS, errs map[string][]string) {
	ds = new(DS)
	errs = make(map[string][]string)

	if len(r.FormValue(labelPrefix+"-digest")) == 0 {
		return
	}

	if keyTag, err := strconv.Atoi(r.FormValue(labelPrefix + "-keytag")); err == nil {
		ds.KeyTag = keyTag
	} else {
		errs[labelPrefix+"-keytag"] = append(errs[labelPrefix+"-keytag"], "Must be a number!")
	}

	if algorithm, err := strconv.Atoi(r.FormValue(labelPrefix + "-algorithm")); err == nil {
		ds.Algorithm = algorithm
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
