// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// Context selects how ambiguous phrases resolve in time.
type Context int

const (
	// ContextFuture (the default) reads an ambiguous phrase as the next
	// occurrence in the future.
	ContextFuture Context = iota
	// ContextPast reads an ambiguous phrase as the most recent past occurrence.
	ContextPast
)

func (c Context) sym() string {
	if c == ContextPast {
		return "past"
	}
	return "future"
}

// Guess selects whether Parse returns a single instant or the whole matched
// [Span], and — for an instant — which point of the span to pick.
type Guess int

const (
	// GuessMiddle (the default) returns the midpoint of the matched span.
	GuessMiddle Guess = iota
	// GuessBegin returns the span's start.
	GuessBegin
	// GuessEnd returns the span's end.
	GuessEnd
	// GuessNone disables guessing; Parse then returns the Span via ParseSpan.
	GuessNone
)

// Options configures a parse, mirroring Chronic::Parser's options hash. The zero
// value is valid and matches the gem's defaults (context future, guess middle,
// sunday week start, ambiguous-time-range 6, endian precedence middle-then-
// little, ambiguous-year-future-bias 50). Now defaults to time.Now.
type Options struct {
	// Now is the anchor instant. Leave zero to use the current time. Pin it for
	// deterministic parses.
	Now time.Time
	// Context resolves ambiguity toward the future (default) or past.
	Context Context
	// Guess selects the returned instant within the matched span.
	Guess Guess
	// AmbiguousTimeRange is the hour (1..12) an ambiguous bare time like "5" is
	// assumed to fall within (AM..PM). Zero means the default of 6. Set
	// AmbiguousTimeRangeNone to disable the assumption entirely.
	AmbiguousTimeRange     int
	AmbiguousTimeRangeNone bool
	// EndianLittle parses "03/04/2011" as day/month (d/m) rather than the
	// default month/day (m/d).
	EndianLittle bool
	// WeekStart is the first day of the week ("sunday" default, or "monday").
	WeekStart string
	// Hours24 forces 24-hour interpretation when true; set Hours24Set together
	// with Hours24=false to force 12-hour interpretation. When Hours24Set is
	// false the parser auto-detects (the gem's nil).
	Hours24    bool
	Hours24Set bool
	// AmbiguousYearFutureBias biases two-digit year expansion (default 50).
	AmbiguousYearFutureBias int
}

// options is the resolved, internal form with defaults applied.
type options struct {
	now                     time.Time
	context                 string
	guess                   Guess
	ambiguousTimeRange      int // -1 == :none
	endianLittle            bool
	weekStart               string
	hours24                 bool
	hours24Set              bool
	ambiguousYearFutureBias int
	text                    string
}

func (o Options) resolve() *options {
	r := &options{
		now:                     o.Now,
		context:                 o.Context.sym(),
		guess:                   o.Guess,
		endianLittle:            o.EndianLittle,
		weekStart:               o.WeekStart,
		hours24:                 o.Hours24,
		hours24Set:              o.Hours24Set,
		ambiguousYearFutureBias: o.AmbiguousYearFutureBias,
	}
	if r.now.IsZero() {
		r.now = time.Now().In(TimeLoc)
	} else {
		r.now = r.now.In(TimeLoc)
	}
	if r.weekStart == "" {
		r.weekStart = "sunday"
	}
	if r.ambiguousYearFutureBias == 0 {
		r.ambiguousYearFutureBias = 50
	}
	switch {
	case o.AmbiguousTimeRangeNone:
		r.ambiguousTimeRange = -1
	case o.AmbiguousTimeRange == 0:
		r.ambiguousTimeRange = 6
	default:
		r.ambiguousTimeRange = o.AmbiguousTimeRange
	}
	return r
}

// Parse parses a natural-language date/time string, returning the chosen instant
// and true, or the zero Time and false when nothing could be parsed. It mirrors
// Chronic.parse with the guess applied.
func Parse(text string, opts ...Options) (time.Time, bool) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Guess == GuessNone {
		// GuessNone via Parse still returns the span's begin (guess(false) returns
		// the span itself in Ruby, but the Time-returning API yields Begin).
		s, ok := ParseSpan(text, o)
		if !ok {
			return time.Time{}, false
		}
		return s.Begin, true
	}
	ro := o.resolve()
	span := parseToSpan(text, ro)
	if span == nil {
		return time.Time{}, false
	}
	return guessInstant(span, o.Guess), true
}

// ParseSpan parses text and returns the whole matched [Span] (guess disabled),
// or nil and false when nothing parsed.
func ParseSpan(text string, opts ...Options) (*Span, bool) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	ro := o.resolve()
	span := parseToSpan(text, ro)
	if span == nil {
		return nil, false
	}
	return span, true
}

// guessInstant ports Parser#guess.
func guessInstant(span *Span, mode Guess) time.Time {
	switch mode {
	case GuessEnd:
		return span.End
	case GuessBegin:
		return span.Begin
	default: // middle
		if span.Width() > 1 {
			return span.Begin.Add(span.End.Sub(span.Begin) / 2)
		}
		return span.Begin
	}
}
