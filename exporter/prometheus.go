package exporter

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	prometheusPath = "/metrics"
	prometheusPort = "9090"
)

type prometheusMetric interface{}

type metrics map[string]prometheusMetric

func register(metrics map[string]prometheusMetric) {
	for _, m := range metrics {
		switch t := m.(type) {
		case *prometheus.CounterVec:
			prometheus.Register(t)
		case prometheus.Counter:
			prometheus.Register(t)
		}
	}
}

func SetPrometheusEndpointAndPort(path, port string) {
	prometheusPath = path
	prometheusPort = port
}

func startPrometheus() {
	go func() {
		http.Handle(prometheusPath, promhttp.Handler())
		log.Fatal(http.ListenAndServe(":"+prometheusPort, nil))
	}()
}
