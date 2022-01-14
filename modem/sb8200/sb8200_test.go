// Copyright 2020 Google Inc. All Rights Reserved.
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

package sb8200

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

	p := "testdata/SB8200.html"
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
			"29": {
				Modulation:    "QAM256",
				Frequency:     "639000000",
				PowerLevel:    1.5,
				SNR:           39.4,
				Correctable:   1643,
				Uncorrectable: 3047,
			},
			"1": {
				Modulation:    "QAM256",
				Frequency:     "459000000",
				PowerLevel:    2.4,
				SNR:           40.1,
				Correctable:   2549,
				Uncorrectable: 8191,
			},
			"2": {
				Modulation:    "QAM256",
				Frequency:     "465000000",
				PowerLevel:    2.8,
				SNR:           40.4,
				Correctable:   2540,
				Uncorrectable: 7489,
			},
			"3": {
				Modulation:    "QAM256",
				Frequency:     "471000000",
				PowerLevel:    2.5,
				SNR:           40.4,
				Correctable:   2505,
				Uncorrectable: 7854,
			},
			"4": {
				Modulation:    "QAM256",
				Frequency:     "477000000",
				PowerLevel:    2.6,
				SNR:           40.5,
				Correctable:   2343,
				Uncorrectable: 7456,
			},
			"5": {
				Modulation:    "QAM256",
				Frequency:     "483000000",
				PowerLevel:    2.1,
				SNR:           40.2,
				Correctable:   2089,
				Uncorrectable: 6803,
			},
			"6": {
				Modulation:    "QAM256",
				Frequency:     "489000000",
				PowerLevel:    1.7,
				SNR:           40.0,
				Correctable:   2092,
				Uncorrectable: 6111,
			},
			"7": {
				Modulation:    "QAM256",
				Frequency:     "495000000",
				PowerLevel:    1.6,
				SNR:           39.9,
				Correctable:   2220,
				Uncorrectable: 5516,
			},
			"8": {
				Modulation:    "QAM256",
				Frequency:     "507000000",
				PowerLevel:    0.5,
				SNR:           39.1,
				Correctable:   2117,
				Uncorrectable: 5893,
			},
			"9": {
				Modulation:    "QAM256",
				Frequency:     "513000000",
				PowerLevel:    0.4,
				SNR:           38.8,
				Correctable:   2210,
				Uncorrectable: 5966,
			},
			"10": {
				Modulation:    "QAM256",
				Frequency:     "519000000",
				PowerLevel:    0.5,
				SNR:           39.4,
				Correctable:   2145,
				Uncorrectable: 5962,
			},
			"11": {
				Modulation:    "QAM256",
				Frequency:     "525000000",
				PowerLevel:    0.4,
				SNR:           39.5,
				Correctable:   1838,
				Uncorrectable: 5681,
			},
			"12": {
				Modulation:    "QAM256",
				Frequency:     "531000000",
				PowerLevel:    0.4,
				SNR:           39.5,
				Correctable:   1760,
				Uncorrectable: 5062,
			},
			"13": {
				Modulation:    "QAM256",
				Frequency:     "543000000",
				PowerLevel:    0.2,
				SNR:           39.5,
				Correctable:   1711,
				Uncorrectable: 4013,
			},
			"14": {
				Modulation:    "QAM256",
				Frequency:     "549000000",
				PowerLevel:    -0.3,
				SNR:           39.0,
				Correctable:   1797,
				Uncorrectable: 3586,
			},
			"15": {
				Modulation:    "QAM256",
				Frequency:     "555000000",
				PowerLevel:    -0.1,
				SNR:           39.1,
				Correctable:   1961,
				Uncorrectable: 3673,
			},
			"16": {
				Modulation:    "QAM256",
				Frequency:     "561000000",
				PowerLevel:    -0.3,
				SNR:           39.0,
				Correctable:   1760,
				Uncorrectable: 4294,
			},
			"17": {
				Modulation:    "QAM256",
				Frequency:     "567000000",
				PowerLevel:    -0.1,
				SNR:           39.0,
				Correctable:   1739,
				Uncorrectable: 4569,
			},
			"18": {
				Modulation:    "QAM256",
				Frequency:     "573000000",
				PowerLevel:    0.3,
				SNR:           39.1,
				Correctable:   1867,
				Uncorrectable: 4407,
			},
			"19": {
				Modulation:    "QAM256",
				Frequency:     "579000000",
				PowerLevel:    0.6,
				SNR:           39.5,
				Correctable:   1761,
				Uncorrectable: 4156,
			},
			"20": {
				Modulation:    "QAM256",
				Frequency:     "585000000",
				PowerLevel:    0.6,
				SNR:           39.4,
				Correctable:   1700,
				Uncorrectable: 3700,
			},
			"21": {
				Modulation:    "QAM256",
				Frequency:     "591000000",
				PowerLevel:    0.5,
				SNR:           39.2,
				Correctable:   1863,
				Uncorrectable: 3231,
			},
			"22": {
				Modulation:    "QAM256",
				Frequency:     "597000000",
				PowerLevel:    0.8,
				SNR:           39.4,
				Correctable:   1895,
				Uncorrectable: 2905,
			},
			"23": {
				Modulation:    "QAM256",
				Frequency:     "603000000",
				PowerLevel:    0.6,
				SNR:           39.0,
				Correctable:   1836,
				Uncorrectable: 3035,
			},
			"24": {
				Modulation:    "QAM256",
				Frequency:     "609000000",
				PowerLevel:    0.6,
				SNR:           39.3,
				Correctable:   2027,
				Uncorrectable: 3141,
			},
			"25": {
				Modulation:    "QAM256",
				Frequency:     "615000000",
				PowerLevel:    0.4,
				SNR:           39.2,
				Correctable:   1765,
				Uncorrectable: 3784,
			},
			"26": {
				Modulation:    "QAM256",
				Frequency:     "621000000",
				PowerLevel:    0.8,
				SNR:           39.2,
				Correctable:   1928,
				Uncorrectable: 4098,
			},
			"27": {
				Modulation:    "QAM256",
				Frequency:     "627000000",
				PowerLevel:    0.9,
				SNR:           39.2,
				Correctable:   1767,
				Uncorrectable: 4253,
			},
			"28": {
				Modulation:    "QAM256",
				Frequency:     "633000000",
				PowerLevel:    1.3,
				SNR:           39.4,
				Correctable:   1848,
				Uncorrectable: 4298,
			},
			"30": {
				Modulation:    "QAM256",
				Frequency:     "645000000",
				PowerLevel:    1.4,
				SNR:           39.3,
				Correctable:   1521,
				Uncorrectable: 3600,
			},
			"31": {
				Modulation:    "QAM256",
				Frequency:     "651000000",
				PowerLevel:    1.9,
				SNR:           39.6,
				Correctable:   1844,
				Uncorrectable: 3185,
			},
			"32": {
				Modulation:    "QAM256",
				Frequency:     "657000000",
				PowerLevel:    1.6,
				SNR:           39.4,
				Correctable:   1836,
				Uncorrectable: 3219,
			},
			"159": {
				Modulation:    "Other",
				Frequency:     "722000000",
				PowerLevel:    2.8,
				SNR:           36.2,
				Correctable:   1179900627,
				Uncorrectable: 0,
			},
		},
		Upstream: map[modem.Channel]*modem.Upstream{
			"1": {
				Frequency:  "23700000",
				PowerLevel: 42.0,
				Modulation: "SC-QAM Upstream",
				Status:     "Locked",
			},
			"2": {
				Frequency:  "17300000",
				PowerLevel: 42.0,
				Modulation: "SC-QAM Upstream",
				Status:     "Locked",
			},
			"3": {
				Frequency:  "30100000",
				PowerLevel: 41.0,
				Modulation: "SC-QAM Upstream",
				Status:     "Locked",
			},
			"4": {
				Frequency:  "36500000",
				PowerLevel: 39.0,
				Modulation: "SC-QAM Upstream",
				Status:     "Locked",
			},
			"5": {
				Frequency:  "41200000",
				PowerLevel: 41.0,
				Modulation: "SC-QAM Upstream",
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
