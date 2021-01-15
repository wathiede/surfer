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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/golang/glog"
	"golang.org/x/net/html"

	"github.com/wathiede/surfer/htmlutil"
	"github.com/wathiede/surfer/modem"
)

const signalURL = "http://192.168.100.1/"

type sb6183 struct {
	fakeData []byte
}

func (sb6183) Name() string { return "SB6183" }

func isSB6183(b []byte) bool {
	return bytes.Contains(b, []byte(`<span id="thisModelNumberIs">SB6183</span>`))
}

func probe(ctx context.Context, client http.Client, path string) modem.Modem {
	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			glog.Errorf("Failed to read %q: %v", path, err)
			return nil
		}
		if isSB6183(b) {
			m, err := NewFakeData(path)
			if err != nil {
				glog.Errorf("Failed to create fake SB6183: %v", err)
				return nil
			}
			return m
		}
		return nil
	}
	glog.Infof("Probing %q", signalURL)
	rc, err := get(ctx, client)
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
	if isSB6183(b) {
		return New()
	}
	return nil
}

func init() {
	modem.Register(probe)
}

// New returns a modem.Modem that scrapes SB6183 formatted data at the default
// URL.
func New() modem.Modem {
	return &sb6183{}
}

// NewFakeData returns a modem.Modem that will parse SB6183 formatted data
// from the HTML file given in path.
func NewFakeData(path string) (modem.Modem, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &sb6183{fakeData: b}, nil
}

func get(ctx context.Context, client http.Client) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", signalURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Status will return signal data parsed from an HTML status page.  If
// sb.fakeData is not nil, the fake data is parsed.  If it is nil, then an
// HTTP request is made to the default signal URL of a SB6183.
func (sb *sb6183) Status(ctx context.Context, client http.Client) (*modem.Signal, error) {
	if sb.fakeData != nil {
		return parseStatus(bytes.NewReader(sb.fakeData))
	}

	rc, err := get(ctx, client)
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
