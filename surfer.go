// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command surfer scrapes the signal status page of a SB6121 cable modem and
// exports values as prometheus metrics.
package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/golang/groupcache/singleflight"
	"github.com/prometheus/client_golang/prometheus"

	"xinu.tv/surfer/modem"
	"xinu.tv/surfer/modem/sb6121"
)

var (
	port         = flag.Int("port", 6666, "port to listen on when serving prometheus metrics")
	fakeDataPath = flag.String("fake", "", "path to fake HTML data.  (default) fetch over HTTP")

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

	codewordsUnerroredMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "codewords_unerrored",
		Help: "Unerrored codeword count",
	},
		[]string{"channel"},
	)
	codewordsCorrectableMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "codewords_correctable",
		Help: "Correctable codeword count",
	},
		[]string{"channel"},
	)
	codewordsUncorrectableMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "codewords_uncorrectable",
		Help: "Uncorrectable codeword count",
	},
		[]string{"channel"},
	)

	upstreamSymbolRateMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upstream_symbol_rate",
		Help: "Upstream symbol rate in sym/sec",
	},
		[]string{"channel", "frequency_hz", "modulation", "ranging_status"},
	)
	upstreamPowerLevelMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upstream_power_level",
		Help: "Upstream power level reading in dBmV",
	},
		[]string{"channel", "frequency_hz", "modulation", "ranging_status"},
	)

	fetchErrorsMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fetch_errors",
		Help: "Count of errors when fetching metrics from modem.",
	},
	)
)

func init() {
	prometheus.MustRegister(downstreamSNRMetric)
	prometheus.MustRegister(downstreamPowerLevelMetric)
	prometheus.MustRegister(upstreamSymbolRateMetric)
	prometheus.MustRegister(upstreamPowerLevelMetric)
	prometheus.MustRegister(codewordsUnerroredMetric)
	prometheus.MustRegister(codewordsCorrectableMetric)
	prometheus.MustRegister(codewordsUncorrectableMetric)
	prometheus.MustRegister(fetchErrorsMetric)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	// TODO(wathiede): probe and create other cable modems depending on content
	// of /index.html
	var m modem.Modem
	if *fakeDataPath != "" {
		var err error
		m, err = sb6121.NewFakeData(*fakeDataPath)
		if err != nil {
			glog.Exitf("Failed to create sb6121 with %q: %v", *fakeDataPath, err)
		}
	} else {
		m = sb6121.New()
	}

	g := &singleflight.Group{}
	ph := prometheus.Handler()
	// Refresh data every prometheus poll.
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only make one query to the cable modem if concurrent requests come in.
		if _, err := g.Do("get", func() (interface{}, error) {
			s, err := m.Status()
			if err != nil {
				fetchErrorsMetric.Inc()
				return nil, err
			}
			for ch, d := range s.Downstream {
				downstreamSNRMetric.WithLabelValues(string(ch), d.Frequency, d.Modulation).Set(d.SNR)
				downstreamPowerLevelMetric.WithLabelValues(string(ch), d.Frequency, d.Modulation).Set(d.PowerLevel)
				codewordsUnerroredMetric.WithLabelValues(string(ch)).Set(d.Unerrored)
				codewordsCorrectableMetric.WithLabelValues(string(ch)).Set(d.Correctable)
				codewordsUncorrectableMetric.WithLabelValues(string(ch)).Set(d.Uncorrectable)
			}

			for ch, u := range s.Upstream {
				upstreamSymbolRateMetric.WithLabelValues(string(ch), u.Frequency, u.Modulation, u.Status).Set(u.SymbolRate)
				upstreamPowerLevelMetric.WithLabelValues(string(ch), u.Frequency, u.Modulation, u.Status).Set(u.PowerLevel)
			}
			return nil, nil
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ph.ServeHTTP(w, r)
	}))
	glog.Fatalf("Listener returned: %v", http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
