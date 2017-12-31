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
	authLinePrefix = "^(?P<date>[A-Z][a-z]{2}\\s+\\d{1,2}) (?P<time>(\\d{2}:?){3}) (?P<hostname>[a-zA-Z_\\-\\.]+) (?P<processName>[a-zA-Z_\\-]+)(\\[(?P<pid>\\d+)\\])?: "

	authLineParsers = map[string]*regexp.Regexp{
		"pamSessionClosed":       regexp.MustCompile(authLinePrefix + "pam_unix\\(.*\\): session closed for user (?P<username>.*)"),
		"pamSessionOpened":       regexp.MustCompile(authLinePrefix + "pam_unix\\(.*\\): session opened for user (?P<username>.*) by"),
		"noIdentificationString": regexp.MustCompile(authLinePrefix + "Did not receive identification string from (?P<ipAddress>.*)"),
		"connectionClosed":       regexp.MustCompile(authLinePrefix + "Connection closed by (?P<ipAddress>.*) port \\d+"),
		"maxAuthAttempts":        regexp.MustCompile(authLinePrefix + "error: maximum authentication attempts exceeded for invalid user (?P<username>.*) from (?P<ipAddress>.*) port \\d+ .*"),
		"invalidUser":            regexp.MustCompile(authLinePrefix + "Invalid user (?P<username>.*) from (?P<ipAddress>.*)"),
		"userNotAllowed":         regexp.MustCompile(authLinePrefix + "User (?P<username>.*) from (?P<ipAddress>.*) not allowed because not listed in .*"),
	}

	ignoredLineParsers = []*regexp.Regexp{
		regexp.MustCompile(authLinePrefix + "(error: )?Received disconnect from .*"),
		regexp.MustCompile(authLinePrefix + "Disconnected from .*"),
		regexp.MustCompile(authLinePrefix + "input_userauth_request: invalid user .*"),
		regexp.MustCompile(authLinePrefix + "fatal: Unable to negotiate with .*"),
		regexp.MustCompile(authLinePrefix + "Disconnecting: Too many authentication failures .*"),
		regexp.MustCompile(authLinePrefix + "Connection reset by .*"),
		regexp.MustCompile(authLinePrefix + "Bad protocol version identification .*"),
		regexp.MustCompile(authLinePrefix + "New session \\d+ of user .*"),
		regexp.MustCompile(authLinePrefix + "Removed session .*"),
		regexp.MustCompile(authLinePrefix + "Accepted publickey for .*"),
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

func LoadAuthLog(filePath string) (*AuthLog, error) {
	authLog := &AuthLog{filePath: filePath, Debug: debug}

	authLog.SetupMetrics()

	exportersRunning = append(exportersRunning, authLog)

	return authLog, nil
}

func (a *AuthLog) StartExport() error {
	var err error
	a.fileHandler, err = tailFile(a, a.filePath)
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
	for t, re := range authLineParsers {
		if re.MatchString(line.Text) {
			parsedLog.Type = t
			matches = getMatches(line.Text, re)
			continue
		}
	}

	if len(matches) == 0 {
		for _, re := range ignoredLineParsers {
			if re.MatchString(line.Text) {
				return
			}
		}
		log.Println("Unknown log type", line.Text)
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

	addMetric(a, parsedLog.IPAddress)
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
			[]string{"hostname", "type", "user", "internal"},
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

func (a *AuthLog) AddMetrics() {
	isInternal := isInternalIP(a.LastLine.IPAddress)

	a.Metrics["line"].(*prometheus.CounterVec).With(prometheus.Labels{
		"hostname": a.LastLine.Hostname,
		"type":     a.LastLine.Type,
		"user":     a.LastLine.Username,
		"internal": fmt.Sprintf("%t", isInternal),
	}).Inc()

	if a.LastLine.IPAddress != "" && dbPath != "" && !isInternal {
		city, err := GetIpLocationDetails(a.LastLine.IPAddress)
		if err != nil {
			log.Println("Error getting ip location details", err)
		}

		if city.Country.IsoCode != "" {
			a.Metrics["location"].(*prometheus.CounterVec).With(prometheus.Labels{
				"continentCode": city.Continent.Code,
				"continentName": city.Continent.Names["en"],
				"countryCode":   city.Country.IsoCode,
				"countryName":   city.Country.Names["en"],
				"city":          city.City.Names["en"],
			}).Inc()
		}
	}
}
