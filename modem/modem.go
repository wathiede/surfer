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

import (
	"context"
	"net/http"
)

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
	Name() string
	// Fetch the status of the modem using implementation specific means.  The
	// context.Context passed in can be used to set timeouts or cancel
	// in-progress requests.
	Status(context.Context, http.Client) (*Signal, error)
}

// NewFunc is registered to determine if a given Modem is available for
// parsing.
// The ctx is used when making any requests.
// Path is optional, if it is empty, implementations should probe their
// configured URL.  If it is non-empty, the contents of the file should be
// used to determine if it is a status page for the given Modem
// implementation.
// Implementations should return nil if path or the default URL do not
// contain expected results.
type NewFunc func(ctx context.Context, client http.Client, path string) Modem

var modems []NewFunc

// New will walk the list of registered cable modems, and returns an instance
// if any probers return successful.  Nil is returned if no probers succeed.
// The ctx is used when making any requests.
// Path is optional, if it is empty, implementations should probe their
// configured URL.  If it is non-empty, the contents of the file should be
// used to determine if it is a status page for the given Modem
// implementation.
func New(ctx context.Context, client http.Client, path string) Modem {
	// TODO(wathiede): run in parallel and take the first that succeeds?
	for _, f := range modems {
		m := f(ctx, client, path)
		if m != nil {
			return m
		}
	}
	return nil
}

// Register allows Modem implementations to register a function to enable
// autodetect for a given implementation.  It is usually called from a package
// init() for the implementation of a Modem.
func Register(f NewFunc) {
	modems = append(modems, f)
}
