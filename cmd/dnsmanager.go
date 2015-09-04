package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var config = struct {
	Listen struct {
		Address net.IP
		Port    int
	}
	TemplatePath  string   `yaml: "template path"`
	BlockedTLDs   []string `yaml: "blocked TLDs"`
	DynamicUpdate struct {
		BinaryPath string `yaml: "binary path"`
		Port       int
	} `yaml: "dynamic update"`
	Transfer struct {
		Address net.IP
		Port    int
		TSigKey struct {
			Name   string
			Secret string
		} `yaml: "tsig key"`
	}
	DNSCheckPort int `yaml: "dns check port"`
}{}

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

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/domain" || r.RequestURI == "/domain/" {
		CreateDomain(w, r)
	} else if strings.HasPrefix(r.RequestURI, "/domain/") {
		UpdateDomain(w, r)
	} else {
		http.Redirect(w, r, "/domain", http.StatusFound)
	}
}
