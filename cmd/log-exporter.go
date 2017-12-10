package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/bah2830/Log-Exporter/exporter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promPort     = flag.String("port", "9090", "Port for prometheus metrics to listen on")
	promEndpoint = flag.String("endpoint", "/metrics", "Endpoint used for metrics")
	authPath     = flag.String("auth", "", "Path to auth.log")
	geoIPPath    = flag.String("geodb", "", "Path to the geoip mmdb file. If not set geoIP lookups will not be enabled")
	debug        = flag.Bool("debug", false, "Run full scan on test logs file")
)

func main() {
	flag.Parse()

	if *geoIPPath != "" {
		exporter.SetGeoIPPath(*geoIPPath)
	}

	if *authPath == "" {
		log.Fatalln("A auth.log path is required")
	}

	authLog, err := exporter.LoadAuthLog(*authPath, *debug)
	if err != nil {
		panic(err)
	}
	defer authLog.Close()

	go authLog.StartExport()

	go func() {
		http.Handle(*promEndpoint, promhttp.Handler())
		log.Fatal(http.ListenAndServe(":"+*promPort, nil))
	}()

	k := make(chan os.Signal, 2)
	<-k
}
