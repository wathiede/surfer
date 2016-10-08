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

// Package modem defines common interfaces and data structures all modem
// scrapers will support.
package modem

type Downstream struct {
	Correctable float64
	// Hz
	Frequency  string
	Modulation string
	// dBmV
	PowerLevel float64
	// dB
	SNR           float64
	Uncorrectable float64
	Unerrored     float64
	// TODO(wathiede): status?
}

type Upstream struct {
	// Hz
	Frequency string
	// Symbols / second
	SymbolRate float64
	// dBmV
	PowerLevel float64
	Modulation string
	Status     string
}

type Channel string

type Signal struct {
	Downstream map[Channel]*Downstream
	Upstream   map[Channel]*Upstream
}

type Modem interface {
	Status() (*Signal, error)
}
