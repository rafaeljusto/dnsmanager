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
	"strconv"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/handler"
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
			fmt.Printf("missing 'config' argument\n")
			cli.ShowAppHelp(c)
			os.Exit(-1)
		}

		loadConfigFile(configFilename)
		redirectLogOutput()
		writePIDToFile()
		initializeTrama()
		startServer()
	}

	app.Run(os.Args)
}

func loadConfigFile(file string) {
	if err := config.Load(file); err != nil {
		log.Fatalf("exiting after an error loading configuration file. Details: %s\n", err)
	}

	abs, _ := filepath.Abs(file)
	fmt.Printf("using %s as configuration file\n", abs)
}

func redirectLogOutput() {
	logFile := path.Join(config.WebNIC.Home, config.WebNIC.Log)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
}

func writePIDToFile() {
	pid := strconv.Itoa(os.Getpid())
	pidFile := path.Join(config.WebNIC.Home, "webnic.pid")

	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Fatalf("cannot write PID [%s] to file %s\n", pid, pidFile)
	}
}

func initializeTrama() {
	// TODO: Trama recover

	groupSet := trama.NewTemplateGroupSet(nil)
	groupSet.FuncMap = template.FuncMap{
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
	}

	handler.Mux.GlobalTemplates = groupSet
	if err := handler.Mux.ParseTemplates(); err != nil {
		log.Fatalf("exiting after an error loading templates. Details: %s\n", err)
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

	http.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir(assetsPath))))
	http.Handle("/", handler.Mux)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
