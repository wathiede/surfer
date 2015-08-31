// Command surfer scrapes the signal status page of a SB6121 cable modem and
// exports values as prometheus metrics.
package main

import (
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/golang/glog"
	"github.com/golang/groupcache/singleflight"
	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/net/html"
)

const signalURL = "http://192.168.100.1/cmSignalData.htm"

var (
	port = flag.Int("port", 6666, "port to listen on when serving prometheus metrics")

	downstreamSNRMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "downstream_snr",
		Help: "Downstream signal-to-noise ratio in dB",
	},
		[]string{"channel", "frequency_hz", "modulation"},
	)
	downstreamPowerLevelMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "downstream_power_level",
		Help: "Downstream power level reading in dBmV",
	},
		[]string{"channel", "frequency_hz", "modulation"},
	)

	upstreamSymbolRateMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upstream_symbol_rate",
		Help: "Upstream symbol rate in sym/sec",
	},
		[]string{"channel", "frequency_hz", "modulation", "ranging_service", "ranging_status"},
	)
	upstreamPowerLevelMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upstream_power_level",
		Help: "Upstream power level reading in dBmV",
	},
		[]string{"channel", "frequency_hz", "modulation", "ranging_service", "ranging_status"},
	)
)

func init() {
	prometheus.MustRegister(downstreamSNRMetric)
	prometheus.MustRegister(downstreamPowerLevelMetric)
	prometheus.MustRegister(upstreamSymbolRateMetric)
	prometheus.MustRegister(upstreamPowerLevelMetric)
}

func getText(n *html.Node) string {
	text := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			text = append(text, c.Data)
		default:
			text = append(text, getText(c))
		}
	}

	return strings.TrimSpace(strings.Join(text, ""))
}

func updateDownstream(n *html.Node) {
	glog.Infoln("Updating downstream table")
	type stat struct {
		frequency  string
		snr        float64
		modulation string
		powerLevel float64
	}
	stats := map[string]*stat{}
	var ids []string

	// Remove nested tables
	for _, t := range cascadia.MustCompile("table table").MatchAll(n) {
		t.Parent.RemoveChild(t)
	}

	for row, tr := range cascadia.MustCompile("tr").MatchAll(n)[1:] {
		switch row {
		case 0:
			// ID
			for _, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				id := getText(td)
				ids = append(ids, id)
				stats[id] = &stat{}
			}
		case 1:
			// Frequency
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].frequency = strings.Fields(getText(td))[0]
			}
		case 2:
			// SNR
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(getText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].snr = f
			}
		case 3:
			// Modulation
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].modulation = getText(td)
			}
		case 4:
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				// Power level
				f, err := strconv.ParseFloat(strings.Fields(getText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].powerLevel = f
			}
		default:
			glog.Fatalf("Unhandled %d row in downstream table", row)
		}
	}
	glog.Infof("updateDownstream data:")
	for k, v := range stats {
		glog.Infof("  %v: %v", k, v)
		downstreamSNRMetric.WithLabelValues(k, v.frequency, v.modulation).Set(v.snr)
		downstreamPowerLevelMetric.WithLabelValues(k, v.frequency, v.modulation).Set(v.powerLevel)
	}
}

func updateUpstream(n *html.Node) {
	glog.Infoln("Updating upstream table")
	type stat struct {
		frequency      string
		rangingService string
		rangingStatus  string
		symbolRate     float64
		modulation     string
		powerLevel     float64
	}
	stats := map[string]*stat{}
	var ids []string
	for row, tr := range cascadia.MustCompile("tr").MatchAll(n)[1:] {
		switch row {
		case 0:
			// ID
			for _, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				id := getText(td)
				ids = append(ids, id)
				stats[id] = &stat{}
			}
		case 1:
			// Frequency
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].frequency = strings.Fields(getText(td))[0]
			}
		case 2:
			// Ranging Service ID
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].rangingService = getText(td)
			}
		case 3:
			// Symbol Rate
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(getText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].symbolRate = f * 1000000
			}
		case 4:
			// Power level
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(getText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].powerLevel = f
			}
		case 5:
			// Modulation
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].modulation = strings.Replace(getText(td), "\n", " ", -1)
			}
		case 6:
			// Ranging Status TODO
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].rangingStatus = getText(td)
			}
		default:
			glog.Fatalf("Unhandled %d row in downstream table", row)
		}
	}
	glog.Infof("updateUpstream data:")
	for k, v := range stats {
		glog.Infof("  %v: %v", k, v)
		upstreamSymbolRateMetric.WithLabelValues(k, v.frequency, v.modulation, v.rangingService, v.rangingStatus).Set(v.symbolRate)
		upstreamPowerLevelMetric.WithLabelValues(k, v.frequency, v.modulation, v.rangingService, v.rangingStatus).Set(v.powerLevel)
	}
}

func updateSignalStats(z *html.Tokenizer) {
	glog.Infoln("Updating signal stats table")
}

func get() error {
	resp, err := http.Get(signalURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	n, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	// All top-level tables are immediate descendants of center.  One table has
	// a nested table in a td, which this filter excludes.
	sel := cascadia.MustCompile("center > table")
	for i, t := range sel.MatchAll(n) {
		glog.Infof("Table %d %v", i, t)
		switch i {
		case 0:
			updateDownstream(t)
		case 1:
			updateUpstream(t)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	defer glog.Flush()

	g := &singleflight.Group{}
	ph := prometheus.Handler()
	// Refresh data every prometheus poll.
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only make one query to the cable modem if concurrent requests come in.
		if _, err := g.Do("get", func() (interface{}, error) {
			if err := get(); err != nil {
				return nil, err
			}
			return nil, nil
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ph.ServeHTTP(w, r)
	}))
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
