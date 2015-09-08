package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rafaeljusto/dnsmanager"
	"gopkg.in/yaml.v2"
)

var config = struct {
	Home   string
	Log    string
	Listen struct {
		Address net.IP
		Port    int
	}
	TemplatePath string `yaml: "template path"`
	AssetsPath   string `yaml: "assets path"`
	TSig         dnsmanager.TSigOptions
	DNSManager   dnsmanager.ServiceConfig `yaml:"dns manager"`
}{}

func main() {
	app := cli.NewApp()
	app.Name = "webnic"
	app.Usage = "NIC web form simulation"
	app.Author = "Rafael Dantas Justo"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config,c",
			EnvVar: "WEB_NIC_CONFIG_FILE",
			Usage:  "configuration file",
		},
	}

	app.Action = func(c *cli.Context) {
		configFilename := c.String("config")

		if configFilename == "" {
			fmt.Printf("missing 'config' argument\n")
			cli.ShowAppHelp(c)
			os.Exit(-1)
		}

		loadConfigFile(configFilename)
		redirectLogOutput()
		writePIDToFile()
		initializeTemplates()
		startServer()
	}

	app.Run(os.Args)
}

func loadConfigFile(file string) {
	if err := yaml.Unmarshal([]byte(file), &config); err != nil {
		log.Fatalf("exiting after an error loading configuration file. Details: %s\n", err)
	}

	abs, _ := filepath.Abs(file)
	fmt.Printf("using %s as configuration file\n", abs)
}

func redirectLogOutput() {
	logFile := path.Join(config.Home, config.Log)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
}

func writePIDToFile() {
	pid := strconv.Itoa(os.Getpid())
	pidFile := path.Join(config.Home, "webnic.pid")

	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Fatalf("cannot write PID [%s] to file %s\n", pid, pidFile)
	}
}

func startServer() {
	var host string
	if config.Listen.Address != nil {
		host = config.Listen.Address.String()
	}
	hostAndPort := net.JoinHostPort(host, strconv.Itoa(config.Listen.Port))

	ln, err := net.Listen("tcp", hostAndPort)
	if err != nil {
		log.Fatal(err)
	}

	assetsPath := path.Join(config.Home, config.AssetsPath)

	http.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir(assetsPath))))
	http.HandleFunc("/", ServeHTTP)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service := dnsmanager.NewService(config.DNSManager)

	var templateData TemplateData
	templateData.Errors = make(map[string][]string)

	if r.RequestURI == "/domain" || r.RequestURI == "/domain/" {
		if r.Method == "GET" {
			templateData.Action = "/domain"
			templateData.NewDomain = true

			// For now we are going to show only two nameservers
			for i := 0; i < 2; i++ {
				templateData.Domain.Nameservers = append(templateData.Domain.Nameservers, dnsmanager.Nameserver{})
			}

			// For now we are going to show only two DSs
			for i := 0; i < 2; i++ {
				templateData.Domain.DSSet = append(templateData.Domain.DSSet, dnsmanager.DS{})
			}

		} else if r.Method == "POST" {
			domain, errs := loadForm(r)
			templateData.Domain = domain
			templateData.Errors = errs

			if len(templateData.Errors) == 0 {
				if err := service.Save(domain); err != nil {
					// TODO!
				}

				templateData.Success = true
				templateData.NewDomain = false
				templateData.Action = "/domain/" + domain.FQDN
			}

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	} else if strings.HasPrefix(r.RequestURI, "/domain/") {
		// TODO: Update domain!

	} else {
		http.Redirect(w, r, "/domain", http.StatusFound)
	}

	var err error
	templateData.RegisteredDomains, err = service.Retrieve(&config.TSig)
	if err != nil {
		log.Printf("error retrieving domains: %s", err)
	}

	if err := domainTemplate.ExecuteTemplate(w, "domain.tmpl.html", templateData); err != nil {
		log.Printf("error parsing template: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
