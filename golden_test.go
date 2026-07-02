// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"
)

// goldenJSON is the deterministic corpus of phrase(+options) -> expected instant
// / span, captured once from the real `chronic` gem 0.10.2 under a fixed anchor
// (Wed Aug 16 2006 14:00:00 UTC). These vectors hold with no Ruby present, so
// they alone drive the ruby-free coverage gate on every CI lane.
//
//go:embed testdata/golden.json
var goldenJSON []byte

// goldenAnchor is the fixed `now:` the vectors were generated with.
var goldenAnchor = time.Date(2006, 8, 16, 14, 0, 0, 0, time.UTC)

type goldenValue struct {
	Span  bool   `json:"span"`
	At    string `json:"at"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// goldenOpts mirrors the Ruby option hash keys we vary in the corpus.
type goldenOpts struct {
	Context            string        `json:"context"`
	Guess              interface{}   `json:"guess"`
	EndianPrecedence   []interface{} `json:"endian_precedence"`
	AmbiguousTimeRange interface{}   `json:"ambiguous_time_range"`
}

type goldenVector struct {
	Phrase string       `json:"phrase"`
	Opts   goldenOpts   `json:"opts"`
	Guess  *goldenValue `json:"guess"`
	SpanV  *goldenValue `json:"span"`
}

const goldenLayout = "2006-01-02 15:04:05"

func loadGolden(t *testing.T) []goldenVector {
	t.Helper()
	var v []goldenVector
	if err := json.Unmarshal(goldenJSON, &v); err != nil {
		t.Fatalf("golden.json: %v", err)
	}
	if len(v) == 0 {
		t.Fatal("golden.json empty")
	}
	return v
}

// toOptions translates a vector's options into the Go Options for the guess call.
func (g goldenOpts) toOptions() Options {
	o := Options{Now: goldenAnchor}
	if g.Context == "past" {
		o.Context = ContextPast
	}
	if len(g.EndianPrecedence) > 0 {
		if s, ok := g.EndianPrecedence[0].(string); ok && s == "little" {
			o.EndianLittle = true
		}
	}
	switch gv := g.Guess.(type) {
	case string:
		switch gv {
		case "begin":
			o.Guess = GuessBegin
		case "end":
			o.Guess = GuessEnd
		case "middle":
			o.Guess = GuessMiddle
		}
	}
	if s, ok := g.AmbiguousTimeRange.(string); ok && s == "none" {
		o.AmbiguousTimeRangeNone = true
	}
	return o
}

func TestGoldenGuess(t *testing.T) {
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()

	for _, vec := range loadGolden(t) {
		vec := vec
		t.Run(vec.Phrase+optSuffix(vec.Opts), func(t *testing.T) {
			got, ok := Parse(vec.Phrase, vec.Opts.toOptions())
			if vec.Guess == nil {
				if ok {
					t.Fatalf("Parse(%q) = %v, want nil", vec.Phrase, got)
				}
				return
			}
			if !ok {
				t.Fatalf("Parse(%q) = nil, want %s", vec.Phrase, vec.Guess.At)
			}
			want, _ := time.ParseInLocation(goldenLayout, vec.Guess.At, time.UTC)
			if !got.Equal(want) {
				t.Errorf("Parse(%q) = %s, want %s", vec.Phrase, got.Format(goldenLayout), vec.Guess.At)
			}
		})
	}
}

func TestGoldenSpan(t *testing.T) {
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()

	for _, vec := range loadGolden(t) {
		vec := vec
		t.Run(vec.Phrase+optSuffix(vec.Opts), func(t *testing.T) {
			o := vec.Opts.toOptions()
			o.Guess = GuessNone
			got, ok := ParseSpan(vec.Phrase, o)
			if vec.SpanV == nil {
				if ok {
					t.Fatalf("ParseSpan(%q) = %v, want nil", vec.Phrase, got)
				}
				return
			}
			if !ok {
				t.Fatalf("ParseSpan(%q) = nil, want a span", vec.Phrase)
			}
			wantB, _ := time.ParseInLocation(goldenLayout, vec.SpanV.Begin, time.UTC)
			wantE, _ := time.ParseInLocation(goldenLayout, vec.SpanV.End, time.UTC)
			if !got.Begin.Equal(wantB) || !got.End.Equal(wantE) {
				t.Errorf("ParseSpan(%q) = %s..%s, want %s..%s", vec.Phrase,
					got.Begin.Format(goldenLayout), got.End.Format(goldenLayout),
					vec.SpanV.Begin, vec.SpanV.End)
			}
		})
	}
}

// optSuffix makes subtests with the same phrase but different options unique.
func optSuffix(o goldenOpts) string {
	s := ""
	if o.Context != "" {
		s += "|ctx=" + o.Context
	}
	if len(o.EndianPrecedence) > 0 {
		s += "|endian"
	}
	if gv, ok := o.Guess.(string); ok {
		s += "|guess=" + gv
	}
	if a, ok := o.AmbiguousTimeRange.(string); ok {
		s += "|atr=" + a
	}
	return s
}
