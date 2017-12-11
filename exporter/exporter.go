package exporter

import (
	"strings"

	"github.com/hpcloud/tail"
)

var (
	debug            = false
	excludeIPs       = make([]string, 0)
	exportersRunning []Exporter
)

type Exporter interface {
	StartExport() error
	Close()
	AddMetrics()
}

func tailFile(e Exporter, filePath string) (*tail.Tail, error) {
	var location *tail.SeekInfo
	if !debug {
		location = &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		}
	}

	return tail.TailFile(filePath, tail.Config{
		Follow:   true,
		Location: location,
	})
}

func addMetric(e Exporter, ip string) {
	exluded := false
	for _, IP := range excludeIPs {
		if IP == ip {
			exluded = true
		}
	}

	if !exluded {
		e.AddMetrics()
	}
}

func Shutdown() {
	for _, e := range exportersRunning {
		e.Close()
	}
}

func Start() {
	startPrometheus()

	for _, e := range exportersRunning {
		go e.StartExport()
	}
}

func SetDebugging(flag bool) {
	debug = flag
}

func SetExludeIPs(ips string) {
	excludeIPs = strings.Split(ips, ",")
}
