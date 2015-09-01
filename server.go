package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/miekg/dns"
)

var (
	domainTemplate *template.Template
	isDomain       = regexp.MustCompile("([a-zA-Z0-9]\\.)*([a-zA-Z0-9](\\.)?)")
	isIP           = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}")

	errNotFound = errors.New("No records for domain")
)

type Nameserver struct {
	Name string
	Glue string
}

type DS struct {
	KeyTag     int
	Algorithm  int
	DigestType uint8
	Digest     string
}

type Domain struct {
	Name        string
	Nameservers []*Nameserver
	DSs         []*DS
}

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

func CreateDomain(w http.ResponseWriter, r *http.Request) {
	templateData := TemplateData{Action: "/domain", NewDomain: true}
	templateData.Errors = make(map[string][]string)
	templateData.Subdomains = retrieveDomains()

	if r.Method == "GET" {
		// For now we are going to show only two nameservers
		for i := 0; i < 2; i++ {
			templateData.Domain.Nameservers = append(templateData.Domain.Nameservers, new(Nameserver))
		}

		// For now we are going to show only two DSs
		for i := 0; i < 2; i++ {
			templateData.Domain.DSs = append(templateData.Domain.DSs, new(DS))
		}

		if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
			log.Fatalf("Error parsing template: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
		}

		return
	}

	createUpdate(w, r, templateData)
}

func UpdateDomain(w http.ResponseWriter, r *http.Request) {
	uriParts := strings.Split(r.RequestURI, "/")

	fqdn := uriParts[len(uriParts)-1]
	if len(fqdn) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	subdomains := retrieveDomains()
	domain, found := subdomains[fqdn]

	if !found {
		log.Printf("Domain %s not found!", fqdn)
		http.Error(w, "", http.StatusNotFound)
		return
	}

	templateData := TemplateData{Action: "/domain/" + fqdn}
	templateData.Errors = make(map[string][]string)
	templateData.Domain = *domain
	templateData.Subdomains = subdomains

	if r.Method == "GET" {
		// For now we are going to show only two nameservers
		for i := 0; i < (2 - len(domain.Nameservers)); i++ {
			templateData.Domain.Nameservers = append(templateData.Domain.Nameservers, new(Nameserver))
		}

		// For now we are going to show only two DSs
		for i := 0; i < (2 - len(domain.DSs)); i++ {
			templateData.Domain.DSs = append(templateData.Domain.DSs, new(DS))
		}

		if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
			log.Fatalf("Error parsing template: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
		}

		return
	}

	createUpdate(w, r, templateData)
}

func createUpdate(w http.ResponseWriter, r *http.Request, templateData TemplateData) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	domain, errs := loadForm(r)
	templateData.Domain = *domain

	if domain.Name == "." || domain.Name == "br." || domain.Name == "com." || domain.Name == "music." {
		errs["domain"] = append(errs["domain"], "The domains 'br', 'com' and 'music' are reserved. Please choose another domain!")
	}

	if len(errs) > 0 {
		templateData.Errors = errs

		if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
			log.Fatalf("Error parsing template: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
		}

		return
	}

	errs = checkDelegation(domain)

	if len(errs) > 0 {
		templateData.Errors = errs

		if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
			log.Fatalf("Error parsing template: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
		}

		return
	}

	if updateRootZone(domain) {
		templateData.Success = true
		templateData.NewDomain = false
		templateData.Action = "/domain/" + domain.Name
		templateData.Subdomains[domain.Name] = domain

	} else {
		templateData.Errors["generic"] = append(templateData.Errors["generic"], "Something went wrong while updating the DNS server")
	}

	if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
		log.Fatalf("Error parsing template: %s", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func updateRootZone(domain *Domain) bool {
	cmdFile, err := ioutil.TempFile("/tmp/", "dnsmanager-")
	if err != nil {
		log.Fatalf("Error creating command file: %s", err.Error())
		return false
	}

	cmdFile.WriteString("update delete " + domain.Name + " NS\n")

	if len(domain.Nameservers[0].Name) > 0 {
		cmdFile.WriteString("update add " + domain.Name + " 172800 NS " + domain.Nameservers[0].Name + "\n")
		cmdFile.WriteString("update delete " + domain.Nameservers[0].Name + " A\n")

		if len(domain.Nameservers[0].Glue) > 0 {
			cmdFile.WriteString("update add " + domain.Nameservers[0].Name + " 172800 A " + domain.Nameservers[0].Glue + "\n")
		}
	}

	if len(domain.Nameservers[1].Name) > 0 {
		cmdFile.WriteString("update add " + domain.Name + " 172800 NS " + domain.Nameservers[1].Name + "\n")
		cmdFile.WriteString("update delete " + domain.Nameservers[1].Name + " A\n")

		if len(domain.Nameservers[1].Glue) > 0 {
			cmdFile.WriteString("update add " + domain.Nameservers[1].Name + " 172800 A " + domain.Nameservers[1].Glue + "\n")
		}
	}

	cmdFile.WriteString("update delete " + domain.Name + " DS\n")

	if len(domain.DSs[0].Digest) > 0 {
		cmdFile.WriteString("update add " + domain.Name + " 172800 DS " + strconv.Itoa(domain.DSs[0].KeyTag) + " " +
			strconv.Itoa(domain.DSs[0].Algorithm) + " " + strconv.FormatUint(uint64(domain.DSs[0].DigestType), 10) + " " + domain.DSs[0].Digest + "\n")
	}

	if len(domain.DSs[1].Digest) > 0 {
		cmdFile.WriteString("update add " + domain.Name + " 172800 DS " + strconv.Itoa(domain.DSs[1].KeyTag) + " " +
			strconv.Itoa(domain.DSs[1].Algorithm) + " " + strconv.FormatUint(uint64(domain.DSs[1].DigestType), 10) + " " + domain.DSs[1].Digest + "\n")
	}

	cmdFile.WriteString("send\n")
	cmdFile.WriteString("quit")
	cmdFile.Close()

	cmd := exec.Command("/usr/local/bind/bin/nsupdate", "-l", "-p", "53", cmdFile.Name())

	var cmdErr bytes.Buffer
	cmd.Stderr = &cmdErr

	if err = cmd.Run(); err != nil {
		log.Printf("Error updating DNS: %s. %s", err.Error(), cmdErr.String())
		return false
	}

	os.Remove(cmdFile.Name())

	return true
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

func retrieveDomains() map[string]*Domain {
	var domains map[string]*Domain
	domains = make(map[string]*Domain)

	client := new(dns.Client)
	client.TsigSecret = map[string]string{"transfer-key.": "zasDqD5nW1USPh4vhLfDBw=="}
	client.Net = "tcp"

	msg := new(dns.Msg)
	msg.SetAxfr(".")
	msg.SetTsig("transfer-key.", dns.HmacMD5, 300, time.Now().Unix())

	transfer := dns.Transfer{
		TsigSecret: map[string]string{"transfer-key.": "zasDqD5nW1USPh4vhLfDBw=="},
	}

	transferChannel, err := transfer.In(msg, "localhost:53")
	if err != nil {
		log.Printf("Error retrieving zones: %s", err.Error())
		return domains
	}

	var glues map[string]string
	glues = make(map[string]string)

	for {
		response, ok := <-transferChannel
		if !ok {
			break
		}

		if response.Error != nil {
			log.Printf("Error retrieving zones: %s", response.Error.Error())
			return domains
		}

		for _, rr := range response.RR {
			if rr.Header().Name == "." || rr.Header().Name == "music." {
				continue
			}

			if rr.Header().Rrtype == dns.TypeNS {
				nsRR := rr.(*dns.NS)

				ns := new(Nameserver)
				ns.Name = nsRR.Ns

				domain := domains[rr.Header().Name]
				if domain == nil {
					domain = new(Domain)
					domain.Name = rr.Header().Name
				}
				domain.Nameservers = append(domain.Nameservers, ns)
				domains[rr.Header().Name] = domain

			} else if rr.Header().Rrtype == dns.TypeDS {
				dsRR := rr.(*dns.DS)

				ds := new(DS)
				ds.KeyTag = int(dsRR.KeyTag)
				ds.Algorithm = int(dsRR.Algorithm)
				ds.DigestType = dsRR.DigestType
				ds.Digest = strings.ToUpper(dsRR.Digest)

				domain := domains[rr.Header().Name]
				if domain == nil {
					domain = new(Domain)
					domain.Name = rr.Header().Name
				}
				domain.DSs = append(domain.DSs, ds)
				domains[rr.Header().Name] = domain

			} else if rr.Header().Rrtype == dns.TypeA {
				aRR := rr.(*dns.A)
				glues[aRR.Header().Name] = aRR.A.String()
			}
		}
	}

	for _, domain := range domains {
		for _, nameserver := range domain.Nameservers {
			for name, glue := range glues {
				if nameserver.Name == name {
					nameserver.Glue = glue
				}
			}
		}
	}

	return domains
}

func checkDelegation(domain *Domain) (errs map[string][]string) {
	if len(domain.Nameservers[0].Name) > 0 {
		errs = checkNS(domain.Nameservers[0], domain.Name, "ns0")
	}

	if len(domain.Nameservers[1].Name) > 0 {
		errs = mergeErrs(errs, checkNS(domain.Nameservers[1], domain.Name, "ns1"))
	}

	if len(errs) > 0 {
		return
	}

	if len(domain.DSs[0].Digest) == 0 && len(domain.DSs[1].Digest) == 0 {
		return
	}

	errs = checkDSs(domain.DSs, domain.Nameservers[0], domain.Name)
	errs = mergeErrs(errs, checkDSs(domain.DSs, domain.Nameservers[1], domain.Name))

	return
}

func checkNS(ns *Nameserver, fqdn, fieldPrefix string) (errs map[string][]string) {
	errs = make(map[string][]string)

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeSOA)
	msg.RecursionDesired = false

	server := ns.Name
	if strings.HasSuffix(ns.Name, fqdn) {
		server = ns.Glue
	}

	client := new(dns.Client)
	response, _, err := client.Exchange(msg, server+":53")

	if err == nil {
		if !response.Authoritative || len(response.Answer) == 0 {
			errs[fieldPrefix] = append(errs[fieldPrefix], "Not authoritative for domain!")
		}

	} else {
		log.Printf("Error while querying %s [%s]. Details: %s", ns.Name, ns.Glue, err.Error())
		errs[fieldPrefix] = append(errs[fieldPrefix], "Fail to query server!")
	}

	return
}

func checkDSs(DSs []*DS, ns *Nameserver, fqdn string) (errs map[string][]string) {
	errs = make(map[string][]string)

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeDNSKEY)
	msg.RecursionDesired = false

	server := ns.Name
	if strings.HasSuffix(ns.Name, fqdn) {
		server = ns.Glue
	}

	client := new(dns.Client)
	response, _, err := client.Exchange(msg, server+":53")

	if err != nil {
		log.Printf("Error while querying DNSKEY in %s [%s]. Details: %s", ns.Name, ns.Glue, err.Error())
		errs["ds0-keytag"] = append(errs["ds0-keytag"], "Fail to query server "+ns.Name+"!")
		errs["ds1-keytag"] = append(errs["ds1-keytag"], "Fail to query server "+ns.Name+"!")
		return
	}

	if response.Truncated {
		client.Net = "tcp"
		response, _, err = client.Exchange(msg, server+":53")

		if err != nil {
			log.Printf("Error while querying DNSKEY in %s [%s]. Details: %s", ns.Name, ns.Glue, err.Error())
			errs["ds0-keytag"] = append(errs["ds0-keytag"], "Fail to query server "+ns.Name+"!")
			errs["ds1-keytag"] = append(errs["ds1-keytag"], "Fail to query server "+ns.Name+"!")
			return
		}
	}

	ds1 := DSs[0]
	ds2 := DSs[1]
	foundDS1 := false
	foundDS2 := false

	for _, rr := range response.Answer {
		if rr.Header().Rrtype == dns.TypeDNSKEY {
			dnskeyRR := rr.(*dns.DNSKEY)

			if int(dnskeyRR.KeyTag()) == ds1.KeyTag {
				foundDS1 = true

				ds := dnskeyRR.ToDS(ds1.DigestType)
				if int(ds.Algorithm) != ds1.Algorithm {
					errs["ds0-algorithm"] = append(errs["ds0-algorithm"], "Algorithm does not match with DNSKEY in "+ns.Name)
				} else if strings.ToUpper(ds.Digest) != ds1.Digest {
					errs["ds0-digest"] = append(errs["ds0-digest"], "Digest does not match with DNSKEY in "+ns.Name)
				} else if dnskeyRR.Flags&dns.SEP == 0 {
					errs["ds0-keytag"] = append(errs["ds0-keytag"], "DNSKEY is not a SEP in "+ns.Name)
				}

			}

			if int(dnskeyRR.KeyTag()) == ds2.KeyTag {
				foundDS2 = true

				ds := dnskeyRR.ToDS(ds2.DigestType)
				if int(ds.Algorithm) != ds2.Algorithm {
					errs["ds1-algorithm"] = append(errs["ds1-algorithm"], "Algorithm does not match with DNSKEY in "+ns.Name)
				} else if strings.ToUpper(ds.Digest) != ds2.Digest {
					errs["ds1-digest"] = append(errs["ds1-digest"], "Digest does not match with DNSKEY in "+ns.Name)
				} else if dnskeyRR.Flags&dns.SEP == 0 {
					errs["ds1-keytag"] = append(errs["ds1-keytag"], "DNSKEY is not a SEP in "+ns.Name)
				}

			}
		}
	}

	if !foundDS1 && len(ds1.Digest) > 0 {
		errs["ds0-keytag"] = append(errs["ds0-keytag"], "Related DNSKEY not found in "+ns.Name)
	}

	if !foundDS2 && len(ds2.Digest) > 0 {
		errs["ds1-keytag"] = append(errs["ds1-keytag"], "Related DNSKEY not found in "+ns.Name)
	}

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

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/domain" || r.RequestURI == "/domain/" {
		CreateDomain(w, r)
	} else if strings.HasPrefix(r.RequestURI, "/domain/") {
		UpdateDomain(w, r)
	} else {
		http.Redirect(w, r, "/domain", http.StatusFound)
	}
}

func redirectLogOutput(logFile string) {
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
}

func writePIDToFile(pidFile string) {
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Print(err)
		log.Printf("Cannot write PID [%s] to file %s. But I don't care\n", pid, pidFile)
	}
}

func main() {
	ipAndPort := ":80"

	ln, err := net.Listen("tcp", ipAndPort)
	if err != nil {
		log.Fatal(err)
	}

	writePIDToFile("server.pid")
	redirectLogOutput("server.log")

	http.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))
	http.HandleFunc("/", ServeHTTP)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}
