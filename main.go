// Copyright 2020 Transnano
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "proftpd" // For Prometheus metrics.
)

var (
	listeningAddress = kingpin.Flag("telemetry.address", "Address on which to expose metrics.").Default(":9064").String()
	metricsEndpoint  = kingpin.Flag("telemetry.endpoint", "Path under which to expose metrics.").Default("/metrics").String()
	// commandPath      = kingpin.Flag("command.path", "Specify the full path to proftpd's ftpwho which show current process information for each FTP session.").Default("/usr/bin/ftpwho").String()
	configFile = kingpin.Flag("config.file", "Specify the full path to proftpd's configuration file.").Default("/etc/proftpd.conf").String()
	// configOption     = kingpin.Flag("config.option", "Option to specify configuration file path. Short or full can be specified.").Default("-c").String()
	// verbose          = kingpin.Flag("verbose", "Enable verbose mode. Output additional information for each connection.").Default("false").Bool()
	gracefulStop = make(chan os.Signal)
)

type Exporter struct {
	FilePath string
	mutex    sync.Mutex

	up             *prometheus.Desc
	uptime         *prometheus.Desc
	users          prometheus.Gauge
	scrapeFailures prometheus.Counter
	connections    *prometheus.GaugeVec
}

func NewExporter(filepath string) *Exporter {
	return &Exporter{
		FilePath: filepath,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the proftpd daemon be reached",
			nil,
			nil),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping proftpd.",
		}),
		users: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "users",
			Help:      "The current user count.",
		}),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime_seconds_total"),
			"Current uptime in seconds (*)",
			nil,
			nil),
		connections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connections",
			Help:      "ProFTPD connection statuses",
		},
			[]string{"state"},
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime
	e.users.Describe(ch)
	e.scrapeFailures.Describe(ch)
	e.connections.Describe(ch)
}

//
func extractNumber(s, sep string) float64 {

	ss := strings.Split(s, sep)
	val, err := strconv.ParseFloat(ss[len(ss)-2], 64)
	if err != nil {
		// FIXME return err??
		return 0
	}

	return val
}

var unitMap = map[string]float64{
	"days": float64(time.Hour * 24 / time.Second),
	"hrs":  float64(time.Hour / time.Second),
	"min":  float64(time.Minute / time.Second),
}

// Calc uptime
func calculateUptime(s string) float64 {
	pattern := `standalone FTP daemon \[(?P<pid>\d*)\], up for(\s(?P<days>\d*) days,)?(\s{1,2}(?P<hrs>\d*) hrs)? (?P<min>\d*) min`
	pathMetadata := regexp.MustCompile(pattern)

	matches := pathMetadata.FindStringSubmatch(s)
	names := pathMetadata.SubexpNames()
	var d float64
	for i, match := range matches {
		if i != 0 {
			if unit, ok := unitMap[names[i]]; ok {
				m := match
				if len(match) == 0 {
					m = "0"
				}
				v, _ := strconv.ParseFloat(m, 64)
				log.Debugf("ff %.0f", v)
				v *= unit
				d += v
				log.Debugf(names[i], m, v)
			}
		}
	}
	return d
}

// Parse uptime
func parseUptime(s string) time.Duration {
	hours, err := time.ParseDuration(s)
	if err != nil {
		hours, _ = time.ParseDuration(s + "m")
	}
	log.Debugf("There are %.0f seconds in %v.\n", hours.Seconds(), hours)
	return hours
}

var commandMap = map[string]string{
	"RETR":         "downloading",
	"READ":         "downloading",
	"scp download": "downloading",
	"STOR":         "uploading",
	"STOU":         "uploading",
	"APPE":         "uploading",
	"WRITE":        "uploading",
	"scp upload":   "uploading",
}

func (e *Exporter) initConnections() {
	e.connections.Reset()
	for k := range commandMap {
		e.connections.WithLabelValues(k)
	}
}

type session struct {
	PID          int64
	User         string
	BeganSession time.Duration
	BeganIdle    time.Duration
	Progress     int64
	Command      string
	Argument     string
}

func check(p, path string) session {
	pathMetadata := regexp.MustCompile(p)

	matches := pathMetadata.FindStringSubmatch(path)
	names := pathMetadata.SubexpNames()

	s := session{}
	for i, match := range matches {
		if i != 0 && len(names[i]) != 0 {
			match = strings.TrimSpace(match)
			fmt.Println(names[i], match)
			switch {
			case "pid" == names[i]:
				val, _ := strconv.ParseInt(match, 10, 64)
				s.PID = val
			case "user" == names[i]:
				s.User = match
			case "began_session" == names[i]:
				s.BeganSession = parseUptime(match)
			case "began_idle" == names[i]:
				s.BeganIdle = parseUptime(match)
			case "progress" == names[i]:
				val, _ := strconv.ParseInt(match, 10, 64)
				s.Progress = val
			case "cmd" == names[i]:
				s.Command = match
			case "arg" == names[i]:
				s.Argument = match
			}
		}
	}

	return s
}

func (e *Exporter) updateConnections(s string) {
	var pattern string
	if strings.HasSuffix(s, "idle") {
		pattern = `\s*(?P<pid>\d+) (?P<user>.+)\s+\[(?P<began_session>.+)\]\s+(?P<began_idle>.+)\s+(?P<cmd>.*)`
	} else if strings.HasSuffix(s, "(authenticating)") {
		pattern = `\s*(?P<pid>\d+) (?P<user>.+)\s+\[(?P<began_session>.+)\]\s+(?P<cmd>.*)`
	} else if strings.Contains(s, " (n/a) ") {
		pattern = `\s*(?P<pid>\d+) (?P<user>.+)\s+\[(?P<began_session>.+)\]\s+\(n/a\)\s+(?P<cmd>.*)\s(?P<arg>.+)`
	} else {
		pattern = `\s*(?P<pid>\d+) (?P<user>.+)\s+\[(?P<began_session>.+)\]\s+\((?P<progress>\d+)%\)\s+(?P<cmd>.*)\s(?P<arg>.+)`
	}
	val := check(pattern, s)
	log.Debugf("%+v\n", val)

	e.connections.WithLabelValues(val.Command).Inc()
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	stdOut := `standalone FTP daemon [15768], up for 78 days, 28 min
30603 account001 [ 1h42m]   0m5s idle
30753 account002 [  0m5s]   0m0s idle
30807 account003 [  0m0s]   0m0s idle
30739 account004 [ 0m10s]   0m0s idle
29525 account005 [ 3h32m]  0m20s idle
29467 account006 [ 1m55s]  1m52s idle
18985 account007 [ 6m45s]  4m50s idle
24309 account008 [ 5h54m]  0m24s idle
11714 account009 [1677h1]   0m6s idle
25929 account010 [ 3m40s]  3m39s idle
30808 account011 [  0m0s]   0m0s idle
32532 account012 [14m12s]  9m28s idle
30748 account013 [  0m8s]   0m8s idle
28379 account014 [ 2m28s] (n/a) STOR img.zip
28379 account015 [ 2m28s] (90%) RETR img.zip
30518 account016 [ 0m19s]  0m19s idle
30857 account017 [32m23s]  0m43s idle
29713 account018 [ 1m41s]  1m41s idle
29714 account019 [ 1m41s]  1m41s idle
 6388 account020 [13m12s]  9m59s idle
 3251 (none)   [  0m3s] (authenticating)
18232 account021 [ 7m20s]  3m39s idle
Service class                      -  21 users`

	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	lines := strings.Split(stdOut, "\n")

	connectionInfo := false

	re := regexp.MustCompile(`^\d+ `)

	e.initConnections()
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "standalone FTP daemon"):
			val := calculateUptime(l)

			ch <- prometheus.MustNewConstMetric(e.uptime, prometheus.CounterValue, val)
		case strings.HasPrefix(l, "Service class"):
			val := extractNumber(l, " ")

			e.users.Set(val)
		case "no users connected" == l:

			e.users.Set(0)
			connectionInfo = false
		case re.MatchString(strings.TrimSpace(l)):
			connectionInfo = true
			e.updateConnections(l)
			// e.scoreboard.Collect(ch)
		}

	}

	e.users.Collect(ch)
	// e.workers.Collect(ch)
	if connectionInfo {
		e.connections.Collect(ch)
	}

	return nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping proftpd: %s", err)
		e.scrapeFailures.Inc()
		e.scrapeFailures.Collect(ch)
	}
}

func main() {

	// Parse flags
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("proftpd_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// listen to termination signals from the OS
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGHUP)
	signal.Notify(gracefulStop, syscall.SIGQUIT)

	exporter := NewExporter(*configFile)
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("proftpd_exporter"))
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	log.Infoln("Starting proftpd_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infof("Starting Server: %s", *listeningAddress)

	// listener for the termination signals from the OS
	go func() {
		log.Infof("listening and wait for graceful stop")
		sig := <-gracefulStop
		log.Infof("caught sig: %+v. Wait 2 seconds...", sig)
		time.Sleep(2 * time.Second)
		log.Infof("Terminate proftpd-exporter on port: %s", *listeningAddress)
		os.Exit(0)
	}()

	http.Handle(*metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
			 <head><title>ProFTPD Exporter</title></head>
			 <body>
			 <h1>ProFTPD Exporter</h1>
			 <p><a href='` + *metricsEndpoint + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
