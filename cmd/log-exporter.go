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
	promPort      = flag.String("port", "9090", "Port for prometheus metrics to listen on")
	promEndpoint  = flag.String("endpoint", "/metrics", "Endpoint used for metrics")
	authPath      = flag.String("auth", "", "Path to auth.log")
	requestPath   = flag.String("request", "", "Path to request.log")
	requestParser = flag.String("request.regexMatch", "", "Regex line used to match on each request line")
	excludedIPs   = flag.String("exluded.ips", "", "list of ips to excluded. Useful to remove monitoring services")
	geoIPPath     = flag.String("geodb", "", "Path to the geoip mmdb file. If not set geoIP lookups will not be enabled")
	debug         = flag.Bool("debug", false, "Run full scan on test logs file")
)

func main() {
	flag.Parse()

	if *requestParser == "" {
		*requestParser = "^\\[.* .0000\\] \\[(?P<domain>.*)\\] \\[(?P<ip_address>[0-9\\.]+)\\] \\[(?P<status>\\d{3})\\] \\[(?P<method>\\w+)\\] .*$"
	}

	if *geoIPPath != "" {
		exporter.SetGeoIPPath(*geoIPPath)
	}

	if *debug {
		exporter.EnableDebugging()
	}

	if *excludedIPs != "" {
		exporter.SetExludeIPs(*excludedIPs)
	}

	somethingStarted := false
	if *authPath != "" {
		authLog, err := exporter.LoadAuthLog(*authPath)
		if err != nil {
			panic(err)
		}
		defer authLog.Close()

		go authLog.StartExport()

		somethingStarted = true
	}

	if *requestPath != "" {
		requestLog, err := exporter.LoadRequestLog(*requestPath, *requestParser)
		if err != nil {
			panic(err)
		}
		defer requestLog.Close()

		go requestLog.StartExport()

		somethingStarted = true
	}

	if !somethingStarted {
		log.Fatalln("No exporters were specified to start")
	}

	go func() {
		http.Handle(*promEndpoint, promhttp.Handler())
		log.Fatal(http.ListenAndServe(":"+*promPort, nil))
	}()

	k := make(chan os.Signal, 2)
	<-k
}
