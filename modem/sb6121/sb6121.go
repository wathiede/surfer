// Copyright 2016 Google Inc. All Rights Reserved.
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

// Package sb6121 scrapes status from the Motorola/ARRIS SB6121.
package sb6121

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/golang/glog"
	"golang.org/x/net/html"

	"github.com/wathiede/surfer/htmlutil"
	"github.com/wathiede/surfer/modem"
)

const signalURL = "http://192.168.100.1/cmSignalData.htm"

type downstreamStat struct {
	frequency  string
	snr        float64
	modulation string
	powerLevel float64
}

type downstreamErrorStat struct {
	unerrored     float64
	correctable   float64
	uncorrectable float64
}

type upstreamStat struct {
	frequency      string
	rangingService string
	rangingStatus  string
	symbolRate     float64
	modulation     string
	powerLevel     float64
}

type sb6121 struct {
	fakeData []byte
}

func (sb6121) Name() string { return "SB6121" }

func isSB6121(b []byte) bool {
	return bytes.Contains(b, []byte(`<META content="Microsoft FrontPage 4.0" name=GENERATOR>`))
}

func probe(ctx context.Context, path string) modem.Modem {
	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			glog.Errorf("Failed to read %q: %v", path, err)
			return nil
		}
		if isSB6121(b) {
			m, err := NewFakeData(path)
			if err != nil {
				glog.Errorf("Failed to create fake SB6121: %v", err)
				return nil
			}
			return m
		}
		return nil
	}
	rc, err := get(ctx)
	if err != nil {
		glog.Errorf("Failed to get status page: %v", err)
		return nil
	}
	defer rc.Close()
	b, err := ioutil.ReadAll(io.LimitReader(rc, 1<<20))
	if err != nil {
		glog.Errorf("Failed to read status page: %v", err)
		return nil
	}
	if isSB6121(b) {
		return New()
	}
	return nil
}

func init() {
	modem.Register(probe)
}

// New returns a modem.Modem that scrapes SB6121 formatted data at the default
// URL.
func New() modem.Modem {
	return &sb6121{}
}

// NewFakeData returns a modem.Modem that will parse SB6121 formatted data
// from the HTML file given in path.
func NewFakeData(path string) (modem.Modem, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &sb6121{fakeData: b}, nil
}

func get(ctx context.Context) (io.ReadCloser, error) {
	glog.V(2).Infof("Start Probing %q", signalURL)
	defer glog.V(2).Infof("Done Probing %q", signalURL)
	c := http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(signalURL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Status will return signal data parsed from an HTML status page.  If
// sb.fakeData is not nil, the fake data is parsed.  If it is nil, then an
// HTTP request is made to the default signal URL of a SB6121.
func (sb *sb6121) Status(ctx context.Context) (*modem.Signal, error) {
	if sb.fakeData != nil {
		return parseStatus(bytes.NewReader(sb.fakeData))
	}
	rc, err := get(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return parseStatus(rc)
}

func parseStatus(r io.Reader) (*modem.Signal, error) {
	n, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	signal := &modem.Signal{
		Downstream: map[modem.Channel]*modem.Downstream{},
		Upstream:   map[modem.Channel]*modem.Upstream{},
	}
	// All top-level tables are immediate descendants of center.  One table has
	// a nested table in a td, which this filter excludes.
	sel := cascadia.MustCompile("center > table")
	for i, t := range sel.MatchAll(n) {
		switch i {
		case 0:
			for ch, s := range updateDownstream(t) {
				signal.Downstream[ch] = &modem.Downstream{
					Frequency:  s.frequency,
					SNR:        s.snr,
					Modulation: s.modulation,
					PowerLevel: s.powerLevel,
				}
			}
		case 1:
			for ch, s := range updateUpstream(t) {
				signal.Upstream[ch] = &modem.Upstream{
					Frequency:  s.frequency,
					Status:     s.rangingStatus,
					SymbolRate: s.symbolRate,
					Modulation: s.modulation,
					PowerLevel: s.powerLevel,
				}
			}
		case 2:
			for ch, s := range updateSignalStats(t) {
				d := signal.Downstream[ch]
				d.Unerrored = s.unerrored
				d.Correctable = s.correctable
				d.Unerrored = s.uncorrectable
			}
		}
	}
	return signal, nil
}

func updateDownstream(n *html.Node) map[modem.Channel]*downstreamStat {
	glog.V(2).Infoln("Updating downstream table")
	stats := map[modem.Channel]*downstreamStat{}
	var ids []modem.Channel

	// Remove nested tables
	for _, t := range cascadia.MustCompile("table table").MatchAll(n) {
		t.Parent.RemoveChild(t)
	}

	for row, tr := range cascadia.MustCompile("tr").MatchAll(n)[1:] {
		switch row {
		case 0:
			// ID
			for _, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				id := modem.Channel(htmlutil.GetText(td))
				ids = append(ids, id)
				stats[id] = &downstreamStat{}
			}
		case 1:
			// Frequency
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].frequency = strings.Fields(htmlutil.GetText(td))[0]
			}
		case 2:
			// SNR
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].snr = f
			}
		case 3:
			// Modulation
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].modulation = htmlutil.GetText(td)
			}
		case 4:
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				// Power level
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].powerLevel = f
			}
		default:
			glog.Fatalf("Unhandled %d row in downstream table", row)
		}
	}
	return stats
}

func updateUpstream(n *html.Node) map[modem.Channel]*upstreamStat {
	glog.V(2).Infoln("Updating upstream table")
	stats := map[modem.Channel]*upstreamStat{}
	var ids []modem.Channel
	for row, tr := range cascadia.MustCompile("tr").MatchAll(n)[1:] {
		switch row {
		case 0:
			// ID
			for _, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				id := modem.Channel(htmlutil.GetText(td))
				ids = append(ids, id)
				stats[id] = &upstreamStat{}
			}
		case 1:
			// Frequency
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].frequency = strings.Fields(htmlutil.GetText(td))[0]
			}
		case 2:
			// Ranging Service ID
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].rangingService = htmlutil.GetText(td)
			}
		case 3:
			// Symbol Rate
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].symbolRate = f * 1000000
			}
		case 4:
			// Power level
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].powerLevel = f
			}
		case 5:
			// Modulation
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].modulation = strings.Replace(htmlutil.GetText(td), "\n", " ", -1)
			}
		case 6:
			// Ranging Status
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				stats[ids[i]].rangingStatus = htmlutil.GetText(td)
			}
		default:
			glog.Fatalf("Unhandled %d row in upstream table", row)
		}
	}
	return stats
}

func updateSignalStats(n *html.Node) map[modem.Channel]*downstreamErrorStat {
	glog.V(2).Infoln("Updating signal stats table")
	stats := map[modem.Channel]*downstreamErrorStat{}
	var ids []modem.Channel
	for row, tr := range cascadia.MustCompile("tr").MatchAll(n)[1:] {
		switch row {
		case 0:
			// ID
			for _, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				id := modem.Channel(htmlutil.GetText(td))
				ids = append(ids, id)
				stats[id] = &downstreamErrorStat{}
			}
		case 1:
			// Total Unerrored Codewords
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].unerrored = f
			}
		case 2:
			// Total Correctable Codewords
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].correctable = f
			}
		case 3:
			// Total Uncorrectable Codewords
			for i, td := range cascadia.MustCompile("td").MatchAll(tr)[1:] {
				f, err := strconv.ParseFloat(strings.Fields(htmlutil.GetText(td))[0], 64)
				if err != nil {
					continue
				}
				stats[ids[i]].uncorrectable = f
			}
		default:
			glog.Fatalf("Unhandled %d row in signal stats table", row)
		}
	}
	return stats
}
