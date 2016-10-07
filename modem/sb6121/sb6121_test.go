package sb6121

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"xinu.tv/surfer/modem"
)

func TestParseStatus(t *testing.T) {
	p := "testdata/SB6121-signal.html"
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
			"10": {
				Correctable:   22563,
				Frequency:     "609000000",
				Modulation:    "QAM256",
				PowerLevel:    9,
				SNR:           37,
				Uncorrectable: 0,
				Unerrored:     110946,
			},
			"11": {
				Correctable:   1.492144e+06,
				Frequency:     "615000000",
				Modulation:    "QAM256",
				PowerLevel:    9,
				SNR:           37,
				Uncorrectable: 0,
				Unerrored:     262486,
			},
			"12": {
				Correctable:   19024,
				Frequency:     "621000000",
				Modulation:    "QAM256",
				PowerLevel:    9,
				SNR:           37,
				Uncorrectable: 0,
				Unerrored:     59971,
			},
			"9": {
				Correctable:   21163,
				Frequency:     "603000000",
				Modulation:    "QAM256",
				PowerLevel:    10,
				SNR:           37,
				Uncorrectable: 0,
				Unerrored:     111242,
			},
		},
		Upstream: map[modem.Channel]*modem.Upstream{
			"1": {
				Frequency:  "30100000",
				SymbolRate: 5.12e+06,
				PowerLevel: 48,
				Modulation: "[3] QPSK [3] 64QAM",
				Status:     "Success",
			},
			"2": {
				Frequency:  "36500000",
				SymbolRate: 5.12e+06,
				PowerLevel: 48,
				Modulation: "[3] QPSK [3] 64QAM",
				Status:     "Success",
			},
			"3": {
				Frequency:  "18900000",
				SymbolRate: 2.56e+06,
				PowerLevel: 47,
				Modulation: "[3] QPSK [3] 64QAM",
				Status:     "Success",
			},
			"4": {
				Frequency:  "23700000",
				SymbolRate: 5.12e+06,
				PowerLevel: 47,
				Modulation: "[3] QPSK [3] 64QAM",
				Status:     "Success",
			},
		},
	}

	if !reflect.DeepEqual(want, got) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("Got:\n%s\nWant:\n%s", g, w)
	}
}
