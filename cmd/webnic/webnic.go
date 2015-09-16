package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/registrobr/trama"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy"
	ajaxhandler "github.com/rafaeljusto/dnsmanager/cmd/webnic/ajax/handler"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
	webhandler "github.com/rafaeljusto/dnsmanager/cmd/webnic/web/handler"
)

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
			fmt.Println("error: missing 'config' argument")
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		loadConfigFile(configFilename)
		redirectLogOutput()
		writePIDToFile()
		initializeTrama()
		initializeHandy()
		startServer()
	}

	app.Run(os.Args)
}

func loadConfigFile(file string) {
	if err := config.Load(file); err != nil {
		log.Fatalf("error loading configuration file: %s\n", err)
	}

	abs, _ := filepath.Abs(file)
	fmt.Printf("using %s as configuration file\n", abs)
}

func redirectLogOutput() {
	logFile := path.Join(config.WebNIC.Home, config.WebNIC.Log)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("error opening log:", err)
		os.Exit(1)
	}

	log.SetOutput(f)
}

func writePIDToFile() {
	pid := strconv.Itoa(os.Getpid())
	pidFile := path.Join(config.WebNIC.Home, "webnic.pid")

	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Fatalf("error: cannot write PID [%s] to file %s\n", pid, pidFile)
	}
}

func initializeTrama() {
	webhandler.Mux.Recover = func(r interface{}) {
		const size = 1 << 16
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		log.Printf("Panic detected. Details: %v\n%s", r, buf)
	}

	groupSet := trama.NewTemplateGroupSet(template.FuncMap{
		"hasErrors": func(field string, errors map[string][]string) bool {
			_, exists := errors[field]
			return exists
		},
		"getErrors": func(field string, errors map[string][]string) []string {
			return errors[field]
		},
		"plusplus": func(number int) int {
			number++
			return number
		},
	})

	webhandler.Mux.GlobalTemplates = groupSet
	if err := webhandler.Mux.ParseTemplates(); err != nil {
		log.Fatalf("error loading templates: %s\n", err)
	}
}

func initializeHandy() {
	ajaxhandler.Mux.Recover = func(r interface{}) {
		const size = 1 << 16
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		log.Printf("Panic detected. Details: %v\n%s", r, buf)
	}

	handy.ErrorFunc = func(err error) {
		log.Printf("handy error: %s\n", err)
	}

	handy.NoMatchFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func startServer() {
	var host string
	if config.WebNIC.Listen.Address != nil {
		host = config.WebNIC.Listen.Address.String()
	}
	hostAndPort := net.JoinHostPort(host, strconv.Itoa(config.WebNIC.Listen.Port))

	ln, err := net.Listen("tcp", hostAndPort)
	if err != nil {
		log.Fatal(err)
	}

	assetsPath := path.Join(config.WebNIC.Home, config.WebNIC.AssetsPath)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath))))
	http.Handle("/domains", ajaxhandler.Mux)
	http.Handle("/domain/", ajaxhandler.Mux)
	http.Handle("/", webhandler.Mux)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
