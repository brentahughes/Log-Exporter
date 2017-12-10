package exporter

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	linePrefix = "^(?P<date>[A-Z][a-z]{2}\\s+\\d{1,2}) (?P<time>(\\d{2}:?){3}) (?P<hostname>[a-zA-Z_\\-\\.]+) (?P<processName>[a-zA-Z_\\-]+)(\\[(?P<pid>\\d+)\\])?: "

	lineParsers = map[string]*regexp.Regexp{
		"pamSessionClosed":       regexp.MustCompile(linePrefix + "pam_unix\\(.*\\): session closed for user (?P<username>.*)"),
		"pamSessionOpened":       regexp.MustCompile(linePrefix + "pam_unix\\(.*\\): session opened for user (?P<username>.*) by"),
		"noIdentificationString": regexp.MustCompile(linePrefix + "Did not receive identification string from (?P<ipAddress>.*)"),
		"connectionClosed":       regexp.MustCompile(linePrefix + "Connection closed by (?P<ipAddress>.*) port \\d+"),
		"maxAuthAttempts":        regexp.MustCompile(linePrefix + "error: maximum authentication attempts exceeded for invalid user (?P<username>.*) from (?P<ipAddress>.*) port \\d+ .*"),
		"invalidUser":            regexp.MustCompile(linePrefix + "Invalid user (?P<username>.*) from (?P<ipAddress>.*)"),
	}
)

type AuthLog struct {
	filePath    string
	LastLogTime time.Time
	LastLine    *AuthLogLine
	fileHandler *tail.Tail
	Metrics     metrics
	Debug       bool
}

type AuthLogLine struct {
	Time      time.Time
	Type      string
	Hostname  string
	Username  string
	IPAddress string
	Process   string
	PID       int
	RawLine   string
}

func LoadAuthLog(filePath string, debug bool) (*AuthLog, error) {
	authLog := &AuthLog{filePath: filePath, Debug: debug}

	authLog.SetupMetrics()

	return authLog, nil
}

func (a *AuthLog) StartExport() error {
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

func (a *AuthLog) ParseLine(line *tail.Line) {
	parsedLog := &AuthLogLine{
		Time:    line.Time,
		RawLine: line.Text,
	}

	matches := make(map[string]string)

	// Find the type of log and parse it
	for t, re := range lineParsers {
		if re.MatchString(line.Text) {
			log.Printf("Found log for type %s\n", t)
			parsedLog.Type = t
			matches = getMatches(line.Text, re)
			continue
		}
	}

	if len(matches) == 0 {
		return
	}

	parsedLog.Hostname = matches["hostname"]
	parsedLog.Process = matches["processName"]
	parsedLog.PID, _ = strconv.Atoi(matches["pid"])

	if v, ok := matches["ipAddress"]; ok {
		parsedLog.IPAddress = v
	}

	if v, ok := matches["username"]; ok {
		parsedLog.Username = v
	}

	a.LastLine = parsedLog

	a.LastLine.AddMetric(a.Metrics)
}

func getMatches(line string, re *regexp.Regexp) map[string]string {
	matches := re.FindStringSubmatch(line)
	results := make(map[string]string)

	// Get the basic information out of the log
	for i, name := range re.SubexpNames() {
		if i != 0 && len(matches) > i {
			results[name] = matches[i]
		}
	}

	return results
}

func (a *AuthLog) SetupMetrics() {
	a.Metrics = metrics{
		"line": prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "log_exporter_auth_lines",
				Help: "Number of lines seen in auth file",
			},
			[]string{"hostname", "process", "type", "ip_address", "user"},
		),
		"location": prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "log_exporter_auth_locations",
				Help: "Number of times each location continent/country/city has requested access",
			},
			[]string{"continentCode", "continentName", "countryCode", "countryName", "city"},
		),
	}

	register(a.Metrics)
}

func (a *AuthLog) Close() {
	a.fileHandler.Stop()
	a.fileHandler.Cleanup()

	database.Close()
}

func (a *AuthLog) GetFilePath() string {
	return a.filePath
}

func (a *AuthLogLine) AddMetric(metrics metrics) {
	metrics["line"].(*prometheus.CounterVec).With(prometheus.Labels{
		"hostname":   a.Hostname,
		"process":    a.Process,
		"type":       a.Type,
		"ip_address": a.IPAddress,
		"user":       a.Username,
	}).Inc()

	if a.IPAddress != "" && dbPath != "" {
		city, err := GetIpLocationDetails(a.IPAddress)
		if err != nil {
			log.Println("Error getting ip location details", err)
		}

		fmt.Println(city.Continent.Names["en"], city.Country.Names["en"], city.City.Names["en"])

		metrics["location"].(*prometheus.CounterVec).With(prometheus.Labels{
			"continentCode": city.Continent.Code,
			"continentName": city.Continent.Names["en"],
			"countryCode":   city.Country.IsoCode,
			"countryName":   city.Country.Names["en"],
			"city":          city.City.Names["en"],
		}).Inc()
	}
}
