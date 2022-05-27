package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	apic "github.com/tomvil/phpipam_exporter/client"
	"github.com/tomvil/phpipam_exporter/collectors"
)

const (
	version string = "1.0"
)

var (
	showVersion = flag.Bool("version", false, "Print version and other information about phpipam_exporter")
	listenAddr  = flag.String("web.listen-address", ":9969", "The address to listen on for HTTP requests")
	metricsPath = flag.String("web.metrics-path", "/metrics", "Path under which metrics will be exposed")
	apiAddress  = flag.String("api.address", "http://127.0.0.1:80", "phpIPAM API address")
	apiUsername = flag.String("api.username", "", "phpIPAM API username")
	apiPassword = flag.String("api.password", "", "phpIPAM API password")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println("phpipam_exporter")
		fmt.Println("Version:", version)
		fmt.Println("Author: Tomas Vilemaitis")
		fmt.Println("Metric exporter for phpIPAM")
		os.Exit(0)
	}

	if *apiUsername == "" {
		*apiUsername = os.Getenv("PHPIPAM_USERNAME")
		if *apiUsername == "" {
			log.Errorln(`Please set the phpIPAM API Username!
		API Username can be set with api.username flag or
		by setting PHPIPAM_USERNAME environment variable.`)
			os.Exit(1)
		}
	}

	if *apiPassword == "" {
		*apiPassword = os.Getenv("PHPIPAM_PASSWORD")
		if *apiPassword == "" {
			log.Errorln(`Please set the phpIPAM API Password!
		API Password can be set with api.password flag or
		by setting PHPIPAM_PASSWORD environment variable.`)
			os.Exit(1)
		}
	}

	startServer()
}

func startServer() {
	var landingPage = []byte(`<html>
	<head><title>phpIPAM Exporter (Version ` + version + `)</title></head>
	<body>
	<h1>phpIPAM Exporter</h1>
	<p><a href="` + *metricsPath + `">Metrics</a></p>
	<h2>More information:</h2>
	<p><a href="https://github.com/tomvil/phpipam_exporter">github.com/tomvil/phpipam_exporter</a></p>
	</body>
	</html>`)

	log.Infof("Starting phpIPAM exporter (Version: %s)", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(landingPage); err != nil {
			log.Fatal(err.Error())
		}
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	apiclient, err := apic.NewClient(*apiAddress, *apiUsername, *apiPassword)
	if err != nil {
		log.Errorln("unable to initialize phpIPAM API client: ", err)
		os.Exit(1)
	}

	l := log.New()
	l.Level = log.ErrorLevel

	registry.MustRegister(collectors.NewSubnetsCollector(apiclient))

	promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      l,
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}
