package exporter

import "github.com/prometheus/client_golang/prometheus"

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
