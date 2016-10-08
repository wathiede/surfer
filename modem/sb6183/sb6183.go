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

// Package sb6183 scrapes status from the Motorola/ARRIS SB6183.
package sb6183

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/golang/glog"
	"golang.org/x/net/html"

	"xinu.tv/surfer/htmlutil"
	"xinu.tv/surfer/modem"
)

const signalURL = "http://192.168.100.1/"

type sb6183 struct {
	fakeData []byte
}

// New returns a modem.Modem that scrapes SB6183 formatted data at the default
// URL.
func New() modem.Modem {
	return &sb6183{}
}

// NewFakeData returns a modem.Modem that with parse SB6183 formatted data
// from the HTML file given in path.
func NewFakeData(path string) (modem.Modem, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &sb6183{fakeData: b}, nil
}

// Status will return signal data parsed from an HTML status page.  If
// sb.fakeData is not nil, the fake data is parsed.  If it is nil, then an
// HTTP request is made to the default signal URL of a SB6183.
func (sb *sb6183) Status() (*modem.Signal, error) {
	if sb != nil {
		return parseStatus(bytes.NewReader(sb.fakeData))
	}

	c := http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(signalURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseStatus(resp.Body)
}

func parseStatus(r io.Reader) (*modem.Signal, error) {
	n, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	sel := cascadia.MustCompile("#thisModelNumberIs")
	mn := sel.MatchFirst(n)
	if mn == nil {
		return nil, errors.New("No thisModelNumberIs ID in HTML")
	}
	glog.Infof("Model: %q", htmlutil.GetText(mn))
	tables := cascadia.MustCompile(".simpleTable").MatchAll(n)
	if len(tables) != 3 {
		return nil, fmt.Errorf("Found %d simpleTables, expected 3", len(tables))
	}
	d, err := parseDownstreamTable(tables[1])
	if err != nil {
		return nil, err
	}
	u, err := parseUpstreamTable(tables[2])
	if err != nil {
		return nil, err
	}
	return &modem.Signal{
		Downstream: d,
		Upstream:   u,
	}, nil
}

func parseDownstreamTable(n *html.Node) (map[modem.Channel]*modem.Downstream, error) {
	m := map[modem.Channel]*modem.Downstream{}
	rows := cascadia.MustCompile("tr").MatchAll(n)
	if len(rows) <= 2 {
		return nil, fmt.Errorf("Expected more than 2 row in table, got %d", len(rows))
	}
	for _, row := range rows[2:] {
		d := &modem.Downstream{}
		var ch modem.Channel
		for i, col := range cascadia.MustCompile("td").MatchAll(row) {
			v := htmlutil.GetText(col)
			fv := v
			if idx := strings.Index(v, " "); idx != -1 {
				fv = fv[:idx]
			}
			f, _ := strconv.ParseFloat(fv, 64)
			switch i {
			case 0:
				// Channel
				ch = modem.Channel(v)
			case 1:
				// Lock Status
			case 2:
				// Modulation
				d.Modulation = v
			case 3:
				// Channel ID
			case 4:
				// Frequency (Hz)
				d.Frequency = v
			case 5:
				// Power (dBmV)
				d.PowerLevel = f
			case 6:
				// SNR (dB)
				d.SNR = f
			case 7:
				// Corrected
				d.Correctable = f
			case 8:
				// Uncorrectables
				d.Uncorrectable = f
			default:
				glog.Errorf("Unexpected %dth column in downstream table", i)
			}
		}
		m[ch] = d
	}
	return m, nil
}

func parseUpstreamTable(n *html.Node) (map[modem.Channel]*modem.Upstream, error) {
	m := map[modem.Channel]*modem.Upstream{}
	rows := cascadia.MustCompile("tr").MatchAll(n)
	if len(rows) <= 2 {
		return nil, fmt.Errorf("Expected more than 2 row in table, got %d", len(rows))
	}
	for _, row := range rows[2:] {
		u := &modem.Upstream{}
		var ch modem.Channel
		for i, col := range cascadia.MustCompile("td").MatchAll(row) {
			v := htmlutil.GetText(col)
			fv := v
			if idx := strings.Index(v, " "); idx != -1 {
				fv = fv[:idx]
			}
			f, _ := strconv.ParseFloat(fv, 64)
			switch i {
			case 0:
				// Channel
				ch = modem.Channel(v)
			case 1:
				// Lock Status
				u.Status = v
			case 2:
				// US Channel Type
				u.Modulation = v
			case 3:
				// Channel ID
			case 4:
				// Symbol Rate
				u.SymbolRate = f * 1000
			case 5:
				// Frequency (Hz)
				u.Frequency = v
			case 6:
				// Power (dBmV)
				u.PowerLevel = f
			default:
				glog.Errorf("Unexpected %dth column in downstream table", i)
			}
		}
		m[ch] = u
	}
	return m, nil
}
