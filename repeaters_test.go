// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"testing"
	"time"
)

// These unit tests drive the repeater engines directly, exercising the
// next()/this()/offset()/width() branches (past pointers, cached-state second
// iterations, and the width accessors) that the phrase corpus reaches only
// partially. The anchor is the golden Wed Aug 16 2006 14:00:00.

var repeaterNow = time.Date(2006, 8, 16, 14, 0, 0, 0, time.UTC)

func withUTC(t *testing.T) {
	t.Helper()
	old := TimeLoc
	TimeLoc = time.UTC
	t.Cleanup(func() { TimeLoc = old })
}

func mk[T repeater](r T) T {
	r.setStart(repeaterNow)
	return r
}

func TestRepeaterWidths(t *testing.T) {
	withUTC(t)
	cases := map[string]int{
		"year":      (&repeaterYear{}).width(),
		"month":     (&repeaterMonth{}).width(),
		"week":      (&repeaterWeek{}).width(),
		"weekday":   (&repeaterWeekday{}).width(),
		"weekend":   (&repeaterWeekend{}).width(),
		"fortnight": (&repeaterFortnight{}).width(),
		"day":       (&repeaterDay{}).width(),
		"hour":      (&repeaterHour{}).width(),
		"minute":    (&repeaterMinute{}).width(),
		"second":    (&repeaterSecond{}).width(),
		"season":    (&repeaterSeason{}).width(),
	}
	for name, w := range cases {
		if w <= 0 {
			t.Errorf("%s width = %d, want > 0", name, w)
		}
	}
	// season-name width comes from the embedded repeaterSeason.
	if (&repeaterSeasonName{}).width() <= 0 {
		t.Error("season-name width should be > 0")
	}
}

// TestRepeaterNextIterates verifies each unit repeater advances on a cached
// second call for both future and past pointers.
func TestRepeaterNextIterates(t *testing.T) {
	withUTC(t)
	build := []struct {
		name string
		new  func() repeater
	}{
		{"year", func() repeater { return &repeaterYear{} }},
		{"month", func() repeater { return &repeaterMonth{} }},
		{"week", func() repeater { return &repeaterWeek{} }},
		{"weekday", func() repeater { return &repeaterWeekday{} }},
		{"weekend", func() repeater { return &repeaterWeekend{} }},
		{"fortnight", func() repeater { return &repeaterFortnight{} }},
		{"day", func() repeater { return &repeaterDay{} }},
		{"hour", func() repeater { return &repeaterHour{} }},
		{"minute", func() repeater { return &repeaterMinute{} }},
		{"second", func() repeater { return &repeaterSecond{} }},
	}
	for _, b := range build {
		for _, ptr := range []string{"future", "past"} {
			r := b.new()
			r.setStart(repeaterNow)
			s1 := r.next(ptr)
			s2 := r.next(ptr) // second call takes the cached-state branch
			if s1 == nil || s2 == nil {
				t.Fatalf("%s/%s: nil span", b.name, ptr)
			}
			if ptr == "future" && !s2.Begin.After(s1.Begin) {
				t.Errorf("%s/future: second next did not advance (%s !> %s)", b.name, s2.Begin, s1.Begin)
			}
			if ptr == "past" && !s2.Begin.Before(s1.Begin) {
				t.Errorf("%s/past: second next did not retreat (%s !< %s)", b.name, s2.Begin, s1.Begin)
			}
		}
	}
}

// TestRepeaterThisPointers exercises the future/past/none arms of this().
func TestRepeaterThisPointers(t *testing.T) {
	withUTC(t)
	build := []func() repeater{
		func() repeater { return &repeaterYear{} },
		func() repeater { return &repeaterMonth{} },
		func() repeater { return &repeaterWeek{} },
		func() repeater { return &repeaterWeekday{} },
		func() repeater { return &repeaterWeekend{} },
		func() repeater { return &repeaterFortnight{} },
		func() repeater { return &repeaterDay{} },
		func() repeater { return &repeaterHour{} },
		func() repeater { return &repeaterMinute{} },
		func() repeater { return &repeaterSecond{} },
		func() repeater { return &repeaterSeason{} },
	}
	for _, nf := range build {
		for _, ptr := range []string{"future", "past", "none"} {
			r := nf()
			r.setStart(repeaterNow)
			if s := r.this(ptr); s == nil {
				t.Errorf("%T.this(%s) = nil", r, ptr)
			}
		}
	}
}

// TestRepeaterOffsets exercises offset() (which drives width via the arrow path)
// for both pointers on every offsetter.
func TestRepeaterOffsets(t *testing.T) {
	withUTC(t)
	span := newSpan(repeaterNow, repeaterNow.Add(time.Second))
	offs := []offsetter{
		&repeaterYear{}, &repeaterMonth{}, &repeaterWeek{}, &repeaterWeekday{},
		&repeaterWeekend{}, &repeaterFortnight{}, &repeaterDay{}, &repeaterHour{},
		&repeaterMinute{}, &repeaterSecond{}, &repeaterSeason{},
	}
	for _, o := range offs {
		if r, ok := o.(repeater); ok {
			r.setStart(repeaterNow)
		}
		if s := o.offset(span, 2, "future"); s == nil {
			t.Errorf("%T.offset(future) = nil", o)
		}
		if s := o.offset(span, 2, "past"); s == nil {
			t.Errorf("%T.offset(past) = nil", o)
		}
	}
}

// TestRepeaterMonthNamePastFuture drives the month-name next arms including the
// cached second iteration (which the corpus reaches only for :none).
func TestRepeaterMonthNamePastFuture(t *testing.T) {
	withUTC(t)
	// past: august on an August anchor resolves to this year's August.
	r := newRepeaterMonthName("august")
	r.setStart(repeaterNow)
	s1 := r.next("past")
	s2 := r.next("past") // cached branch, year-1
	if s2.Begin.Year() != s1.Begin.Year()-1 {
		t.Errorf("month-name past cache: %d then %d", s1.Begin.Year(), s2.Begin.Year())
	}
	// future: a month before the anchor rolls to next year.
	rf := newRepeaterMonthName("january")
	rf.setStart(repeaterNow)
	f1 := rf.next("future")
	f2 := rf.next("future")
	if f2.Begin.Year() != f1.Begin.Year()+1 {
		t.Errorf("month-name future cache: %d then %d", f1.Begin.Year(), f2.Begin.Year())
	}
	// a month after the anchor resolves this year with :future.
	ra := newRepeaterMonthName("december")
	ra.setStart(repeaterNow)
	if got := ra.next("future").Begin.Year(); got != repeaterNow.Year() {
		t.Errorf("december future year = %d", got)
	}
}

// TestRepeaterMonthNameSameMonthPanics: :future/:past on the anchor's own month
// panics with the sentinel (surfaced as a no-parse).
func TestRepeaterMonthNameSameMonthPanics(t *testing.T) {
	withUTC(t)
	defer func() {
		if r := recover(); r != errDayOverflow {
			t.Fatalf("recover = %v, want errDayOverflow", r)
		}
	}()
	r := newRepeaterMonthName("august")
	r.setStart(repeaterNow)
	r.next("future") // now.Month == idx with :future -> panic
}

func TestNewUnitRepeaterTagUnknown(t *testing.T) {
	if newUnitRepeaterTag("bogus", &options{}) != nil {
		t.Error("unknown unit should yield a nil tag")
	}
}

func TestRepeaterTimeTooManyGroups(t *testing.T) {
	// >4 colon groups yields the guarded zero-Tick constructor.
	r := newRepeaterTime("1:2:3:4:5", &options{})
	if r.typ.seconds != 0 {
		t.Errorf("malformed time seconds = %v, want 0", r.typ.seconds)
	}
}

func TestRepeaterTimePastIterates(t *testing.T) {
	withUTC(t)
	// A bare unambiguous 24h time in the past, iterated backwards.
	r := newRepeaterTime("15", &options{hours24: true, hours24Set: true})
	r.setStart(repeaterNow)
	s1 := r.next("past")
	s2 := r.next("past")
	if !s2.Begin.Before(s1.Begin) {
		t.Errorf("time past did not retreat: %s then %s", s1.Begin, s2.Begin)
	}
	// ambiguous time iterates by half-days.
	ra := newRepeaterTime("5", &options{})
	ra.setStart(repeaterNow)
	a1 := ra.next("future")
	a2 := ra.next("future")
	if !a2.Begin.After(a1.Begin) {
		t.Errorf("ambiguous time future did not advance: %s then %s", a1.Begin, a2.Begin)
	}
}

func TestFindCurrentSeasonNone(t *testing.T) {
	withUTC(t)
	r := &repeaterSeason{}
	r.setStart(repeaterNow)
	// A date inside summer classifies as summer.
	if got := r.findCurrentSeason(miniDate{8, 16}); got != "summer" {
		t.Errorf("aug 16 season = %q, want summer", got)
	}
}

func TestParseGenericTimestampBranches(t *testing.T) {
	withUTC(t)
	// space-separated form takes the space->T normalisation branch.
	got, ok := parseGenericTimestamp("2016-05-27 14:30:00")
	if !ok || got.Year() != 2016 || got.Hour() != 14 {
		t.Errorf("space form = %v ok=%v", got, ok)
	}
	// a non-timestamp returns false.
	if _, ok := parseGenericTimestamp("not a date"); ok {
		t.Error("garbage should not parse")
	}
}
