// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"testing"
	"time"
)

// These unit tests exercise the pure helper functions and option-resolution
// branches that the phrase corpus does not reach directly (overflow arithmetic,
// boundary predicates, and option defaults). They need no Ruby.

func TestConstructOverflow(t *testing.T) {
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()

	cases := []struct {
		name string
		args []int
		want time.Time
	}{
		// second >= 60 rolls into minutes.
		{"secroll", []int{2006, 1, 1, 0, 0, 75}, time.Date(2006, 1, 1, 0, 1, 15, 0, time.UTC)},
		// minute >= 60 rolls into hours.
		{"minroll", []int{2006, 1, 1, 0, 75, 0}, time.Date(2006, 1, 1, 1, 15, 0, 0, time.UTC)},
		// hour >= 24 rolls into days.
		{"hourroll", []int{2006, 1, 1, 26, 0, 0}, time.Date(2006, 1, 2, 2, 0, 0, 0, time.UTC)},
		// day > 28 within a leap February overflows into the next month.
		{"dayrollleap", []int{2004, 2, 40}, time.Date(2004, 3, 11, 0, 0, 0, 0, time.UTC)},
		// day > 28 within a non-leap month overflows into the next month.
		{"dayroll", []int{2006, 4, 40}, time.Date(2006, 5, 10, 0, 0, 0, 0, time.UTC)},
		// month exactly a multiple of 12 clamps to December of the prior years.
		{"monthmul12", []int{2006, 24}, time.Date(2007, 12, 1, 0, 0, 0, 0, time.UTC)},
		// month > 12 not a multiple of 12 rolls the year.
		{"monthroll", []int{2006, 14}, time.Date(2007, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := construct(c.args[0], c.args[1:]...)
			if !got.Equal(c.want) {
				t.Errorf("construct(%v) = %s, want %s", c.args, got, c.want)
			}
		})
	}
}

func TestConstructDayOverflowPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r != errDayOverflow {
			t.Fatalf("recover() = %v, want errDayOverflow", r)
		}
	}()
	construct(2006, 1, 99) // day > 56 must panic with the sentinel.
	t.Fatal("construct did not panic")
}

func TestMonthOverflow(t *testing.T) {
	cases := []struct {
		year, month, day int
		want             bool
	}{
		{2006, 0, 1, false},  // month < 1
		{2006, 13, 1, false}, // month > 12
		{2006, 2, 29, true},  // non-leap feb 29 overflows
		{2004, 2, 29, false}, // leap feb 29 ok
		{2004, 2, 30, true},  // leap feb 30 overflows
		{2006, 4, 31, true},  // april has 30 days
		{2006, 1, 31, false}, // january ok
	}
	for _, c := range cases {
		if got := monthOverflow(c.year, c.month, c.day); got != c.want {
			t.Errorf("monthOverflow(%d,%d,%d) = %v, want %v", c.year, c.month, c.day, got, c.want)
		}
	}
}

func TestLastN(t *testing.T) {
	toks := []*token{newToken("a"), newToken("b"), newToken("c")}
	if lastN(toks, 0) != nil {
		t.Error("lastN(_,0) should be nil")
	}
	if lastN(toks, -1) != nil {
		t.Error("lastN(_,-1) should be nil")
	}
	if got := lastN(toks, 5); len(got) != 3 {
		t.Errorf("lastN(_,5) len = %d, want 3", len(got))
	}
	if got := lastN(toks, 2); len(got) != 2 || got[0].word != "b" {
		t.Errorf("lastN(_,2) = %v", got)
	}
}

func TestTimeWithRollover(t *testing.T) {
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()

	p := &parser{now: time.Date(2006, 8, 16, 14, 0, 0, 0, time.UTC)}
	// no overflow: exact date.
	if got := p.timeWithRollover(2006, 8, 17); !got.Equal(time.Date(2006, 8, 17, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("no-overflow = %s", got)
	}
	// mid-month overflow rolls to the 1st of the next month.
	if got := p.timeWithRollover(2006, 4, 31); !got.Equal(time.Date(2006, 5, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("mid-year overflow = %s", got)
	}
	// december overflow rolls to jan 1 of the next year.
	if got := p.timeWithRollover(2006, 12, 32); !got.Equal(time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("december overflow = %s", got)
	}
}

func TestCouldBeHour12(t *testing.T) {
	if !couldBeHour(12, true) {
		t.Error("12 should be a valid 12-hour value")
	}
	if couldBeHour(13, true) {
		t.Error("13 is not a valid 12-hour value")
	}
	if !couldBeHour(24, false) {
		t.Error("24 should be a valid 24-hour value")
	}
}

func TestIsBetweenEdges(t *testing.T) {
	// same start/end/target month with the day outside the [start,end] window.
	md := miniDate{6, 15}
	if md.isBetween(miniDate{6, 20}, miniDate{6, 10}) {
		t.Error("day outside same-month window should be false")
	}
	// target equals the start month and day within range.
	if !(miniDate{6, 25}).isBetween(miniDate{6, 20}, miniDate{9, 22}) {
		t.Error("start-month in range should be true")
	}
	// target month falls in the wrapping ring between start and end.
	if !(miniDate{7, 1}).isBetween(miniDate{6, 20}, miniDate{9, 22}) {
		t.Error("mid-ring month should be true")
	}
	// wrapping season (winter) across the year boundary.
	if !(miniDate{1, 10}).isBetween(miniDate{12, 22}, miniDate{3, 19}) {
		t.Error("january should be within winter")
	}
	// a month entirely outside the ring.
	if (miniDate{7, 1}).isBetween(miniDate{12, 22}, miniDate{3, 19}) {
		t.Error("july is not within winter")
	}
}

func TestIndexOfMissing(t *testing.T) {
	if indexOf([]string{"a", "b"}, "z") != -1 {
		t.Error("missing element should return -1")
	}
}

func TestMakeYearFourDigit(t *testing.T) {
	// already four digits passes through unchanged.
	if got := makeYear(1997, 50, 2006); got != 1997 {
		t.Errorf("makeYear(1997) = %d", got)
	}
	// two-digit year below the start year is pushed forward a century.
	if got := makeYear(1, 50, 2006); got != 2001 {
		t.Errorf("makeYear(01) = %d, want 2001", got)
	}
}

func TestByName(t *testing.T) {
	d := &definitions{}
	d.date = []handler{{}}
	if got := d.byName("date"); len(got) != 1 {
		t.Errorf("byName(date) len = %d", len(got))
	}
	if got := d.byName("bogus"); got != nil {
		t.Errorf("byName(bogus) = %v, want nil", got)
	}
}

func TestUntag(t *testing.T) {
	tok := newToken("x")
	tok.tag(&tag{class: tcSeparatorAt})
	tok.tag(&tag{class: tcScalar, num: 5})
	tok.untag(tcSeparatorAt)
	if tok.getTag(tcSeparatorAt) != nil {
		t.Error("untag left the separator tag")
	}
	if tok.getTag(tcScalar) == nil {
		t.Error("untag removed the scalar tag")
	}
}

func TestScanSignMinus(t *testing.T) {
	toks := []*token{newToken("-"), newToken("+"), newToken("x")}
	scanSign(toks)
	if toks[0].getTag(tcSignMinus) == nil {
		t.Error("- should be tagged minus")
	}
	if toks[1].getTag(tcSignPlus) == nil {
		t.Error("+ should be tagged plus")
	}
	if toks[2].tagged() {
		t.Error("x should be untagged")
	}
}

func TestInsertAt(t *testing.T) {
	got := insertAt([]string{"a", "c"}, 1, "b")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("insertAt = %v", got)
	}
}

func TestParseNoNowUsesTimeNow(t *testing.T) {
	// With no Now, resolve() falls back to time.Now; "today" must parse.
	if _, ok := Parse("today"); !ok {
		t.Error("Parse(today) with default Now should succeed")
	}
	// A bogus phrase returns false.
	if _, ok := Parse("gobbledigook"); ok {
		t.Error("Parse(gobbledigook) should fail")
	}
}

func TestParseGuessNone(t *testing.T) {
	now := time.Date(2006, 8, 16, 14, 0, 0, 0, time.UTC)
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()
	// GuessNone via Parse returns the span's Begin.
	got, ok := Parse("may 2016", Options{Now: now, Guess: GuessNone})
	if !ok {
		t.Fatal("Parse(may 2016, GuessNone) failed")
	}
	want := time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("GuessNone begin = %s, want %s", got, want)
	}
	// GuessNone on an unparseable phrase returns false.
	if _, ok := Parse("gobbledigook", Options{Now: now, Guess: GuessNone}); ok {
		t.Error("GuessNone bogus should fail")
	}
}

func TestParseAmbiguousTimeRangeCustom(t *testing.T) {
	now := time.Date(2006, 8, 16, 14, 0, 0, 0, time.UTC)
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()
	// A non-zero AmbiguousTimeRange takes the default-override branch in resolve.
	if _, ok := Parse("5", Options{Now: now, AmbiguousTimeRange: 8}); !ok {
		t.Error("Parse(5, ATR=8) should parse")
	}
}
