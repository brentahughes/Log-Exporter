package main

import (
	"flag"
	"os"

	"github.com/bah2830/Log-Exporter/exporter"
)

var (
	promPort      = flag.String("promethues.port", "9090", "Port for prometheus metrics to listen on")
	promEndpoint  = flag.String("prometheus.endpoint", "/metrics", "Endpoint used for metrics")
	authPath      = flag.String("auth.path", "", "Path to auth.log")
	requestPath   = flag.String("request.path", "", "Path to request.log")
	requestParser = flag.String("request.regexMatch", "", "Regex line used to match on each request line")
	excludedIPs   = flag.String("exludedIPs", "", "list of ips to excluded. Useful to remove monitoring services")
	geoIPPath     = flag.String("geodb", "", "Path to the geoip mmdb file. If not set geoIP lookups will not be enabled")
	debug         = flag.Bool("debug", false, "Run full scan on test logs file")
)

func main() {
	flag.Parse()

	if *requestParser == "" {
		*requestParser = "^\\[.* .0000\\] \\[(?P<domain>.*)\\] \\[(?P<ip_address>[0-9\\.]+)\\] \\[(?P<status>\\d{3})\\] \\[(?P<method>\\w+)\\] .*$"
	}

	// Setup data for exporters
	exporter.SetDebugging(*debug)
	exporter.SetGeoIPPath(*geoIPPath)
	exporter.SetExludeIPs(*excludedIPs)
	exporter.SetPrometheusEndpointAndPort(*promEndpoint, *promPort)

	if *authPath != "" {
		if _, err := exporter.LoadAuthLog(*authPath); err != nil {
			panic(err)
		}
	}

	if *requestPath != "" {
		if _, err := exporter.LoadRequestLog(*requestPath, *requestParser); err != nil {
			panic(err)
		}
	}

	// Start the file listeners
	exporter.Start()

	// Wait for kill signal
	k := make(chan os.Signal, 2)
	<-k

	// Shutdown all exporters gracefully
	exporter.Shutdown()
}
