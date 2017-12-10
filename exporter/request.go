package exporter

import (
	"log"
	"regexp"
	"time"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
)

type RequestLog struct {
	filePath    string
	logParser   *regexp.Regexp
	LastLogTime time.Time
	LastLine    *RequestLogLine
	fileHandler *tail.Tail
	Metrics     metrics
	Debug       bool
}

type RequestLogLine struct {
	Time       time.Time
	Domain     string
	IPAddress  string
	StatusCode string
	Method     string
	RawLine    string
}

func LoadRequestLog(filePath, lineParser string) (*RequestLog, error) {
	requestLog := &RequestLog{
		filePath:  filePath,
		logParser: regexp.MustCompile(lineParser),
		Debug:     debug,
	}

	requestLog.SetupMetrics()

	return requestLog, nil
}

func (a *RequestLog) StartExport() error {
	var err error
	a.fileHandler, err = tailFile(a, a.Debug)
	if err != nil {
		log.Fatalln(err)
	}

	for line := range a.fileHandler.Lines {
		a.ParseLine(line)
		a.LastLogTime = time.Now().UTC()
	}

	return nil
}

func (a *RequestLog) ParseLine(line *tail.Line) {
	parsedLog := &RequestLogLine{
		Time:    line.Time,
		RawLine: line.Text,
	}

	matches := make(map[string]string)

	if a.logParser.MatchString(line.Text) == false {
		log.Println("Could not parse line", line.Text)
		return
	}

	results := a.logParser.FindStringSubmatch(line.Text)

	// Get the basic information out of the log
	for i, name := range a.logParser.SubexpNames() {
		if i != 0 && len(results) > i {
			matches[name] = results[i]
		}
	}

	if len(matches) == 0 {
		return
	}

	parsedLog.Domain = matches["domain"]
	parsedLog.IPAddress = matches["ip_address"]
	parsedLog.Method = matches["method"]
	parsedLog.StatusCode = matches["status"]
	a.LastLine = parsedLog

	exluded := false
	for _, IP := range excludeIPs {
		if IP == parsedLog.IPAddress {
			exluded = true
		}
	}

	if !exluded {
		a.LastLine.AddMetric(a.Metrics)
	}
}

func (a *RequestLog) SetupMetrics() {
	a.Metrics = metrics{
		"line": prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "log_exporter_request_lines",
				Help: "Number of lines seen in request file",
			},
			[]string{"domain", "method", "status"},
		),
		"location": prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "log_exporter_request_locations",
				Help: "Number of times each location continent/country/city has requested access in the request log",
			},
			[]string{"domain", "continentCode", "continentName", "countryCode", "countryName", "city"},
		),
	}

	register(a.Metrics)
}

func (a *RequestLog) Close() {
	a.fileHandler.Stop()
	a.fileHandler.Cleanup()

	database.Close()
}

func (a *RequestLog) GetFilePath() string {
	return a.filePath
}

func (a *RequestLogLine) AddMetric(metrics metrics) {
	metrics["line"].(*prometheus.CounterVec).With(prometheus.Labels{
		"domain": a.Domain,
		"status": a.StatusCode,
		"method": a.Method,
	}).Inc()

	if a.IPAddress != "" && dbPath != "" {
		city, err := GetIpLocationDetails(a.IPAddress)
		if err != nil {
			log.Println("Error getting ip location details", err)
		}

		if city.Country.IsoCode != "" {
			metrics["location"].(*prometheus.CounterVec).With(prometheus.Labels{
				"domain":        a.Domain,
				"continentCode": city.Continent.Code,
				"continentName": city.Continent.Names["en"],
				"countryCode":   city.Country.IsoCode,
				"countryName":   city.Country.Names["en"],
				"city":          city.City.Names["en"],
			}).Inc()
		} else {
			log.Println("No geo ip data found for", a.IPAddress)
		}

	}
}
