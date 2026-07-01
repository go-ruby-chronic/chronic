// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// Span is a half-open range of time [Begin, End). It mirrors Chronic::Span,
// which subclasses Ruby's Range over Time.
type Span struct {
	Begin time.Time
	End   time.Time
}

// newSpan builds a Span from two instants.
func newSpan(begin, end time.Time) *Span {
	return &Span{Begin: begin, End: end}
}

// Width returns the span's duration in whole seconds.
func (s *Span) Width() int {
	return int(s.End.Sub(s.Begin).Seconds())
}

// addSeconds returns a new Span shifted forward by n seconds.
func (s *Span) addSeconds(n int) *Span {
	d := time.Duration(n) * time.Second
	return &Span{Begin: s.Begin.Add(d), End: s.End.Add(d)}
}

// cover reports whether t lies within [Begin, End], matching Range#cover? for
// the inclusive endpoints Chronic relies on in find_within.
func (s *Span) cover(t time.Time) bool {
	return !t.Before(s.Begin) && !t.After(s.End)
}
