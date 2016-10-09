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

package sb6183

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"

	"github.com/wathiede/surfer/modem"
)

func TestParseStatus(t *testing.T) {
	flag.Set("v", "true")
	flag.Set("logtostderr", "true")

	p := "testdata/SB6183.html"
	r, err := os.Open(p)
	if err != nil {
		t.Fatalf("Failed to open %q: %v", p, err)
	}
	defer r.Close()

	got, err := parseStatus(r)
	if err != nil {
		t.Fatalf("Failed to parse %q: %v", p, err)
	}

	want := &modem.Signal{
		Downstream: map[modem.Channel]*modem.Downstream{
			"1": {
				Correctable:   0,
				Frequency:     "555000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    6.3,
				SNR:           38.4,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"10": {
				Correctable:   0,
				Frequency:     "609000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3.7,
				SNR:           37.1,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"11": {
				Correctable:   3,
				Frequency:     "615000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3.5,
				SNR:           37,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"12": {
				Correctable:   3,
				Frequency:     "621000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3.2,
				SNR:           36.9,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"13": {
				Correctable:   5,
				Frequency:     "627000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3.1,
				SNR:           36.7,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"14": {
				Correctable:   10,
				Frequency:     "633000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3,
				SNR:           36.7,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"15": {
				Correctable:   8,
				Frequency:     "639000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3,
				SNR:           36.6,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"16": {
				Correctable:   7,
				Frequency:     "645000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3,
				SNR:           36.7,
				Uncorrectable: 9,
				Unerrored:     0,
			},
			"2": {
				Correctable:   0,
				Frequency:     "561000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    5.8,
				SNR:           38.4,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"3": {
				Correctable:   0,
				Frequency:     "567000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    5.5,
				SNR:           38.3,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"4": {
				Correctable:   0,
				Frequency:     "573000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    5.5,
				SNR:           38.2,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"5": {
				Correctable:   0,
				Frequency:     "579000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    5.1,
				SNR:           38,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"6": {
				Correctable:   0,
				Frequency:     "585000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    4.8,
				SNR:           37.7,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"7": {
				Correctable:   0,
				Frequency:     "591000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    4.6,
				SNR:           37.5,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"8": {
				Correctable:   0,
				Frequency:     "597000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    4.2,
				SNR:           37.3,
				Uncorrectable: 0,
				Unerrored:     0,
			},
			"9": {
				Correctable:   3,
				Frequency:     "603000000 Hz",
				Modulation:    "QAM256",
				PowerLevel:    3.9,
				SNR:           37.2,
				Uncorrectable: 0,
				Unerrored:     0,
			},
		},
		Upstream: map[modem.Channel]*modem.Upstream{
			"1": {
				Frequency:  "36500000 Hz",
				SymbolRate: 5.12e+06,
				PowerLevel: 36,
				Modulation: "ATDMA",
				Status:     "Locked",
			},
			"2": {
				Frequency:  "30100000 Hz",
				SymbolRate: 5.12e+06,
				PowerLevel: 35.5,
				Modulation: "ATDMA",
				Status:     "Locked",
			},
			"3": {
				Frequency:  "18900000 Hz",
				SymbolRate: 2.56e+06,
				PowerLevel: 33,
				Modulation: "ATDMA",
				Status:     "Locked",
			},
			"4": {
				Frequency:  "23700000 Hz",
				SymbolRate: 5.12e+06,
				PowerLevel: 33.5,
				Modulation: "ATDMA",
				Status:     "Locked",
			},
		},
	}

	if !reflect.DeepEqual(want, got) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("Got:\n%s\nWant:\n%s", g, w)
	}
}
