// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"testing"
	"time"
)

// This file covers the residual error/edge branches that neither the phrase
// corpus nor the engine tests reach: handler nil-returns (a bad anchor, a
// non-offsetter repeater, day overflow), the numerizer's big-prefix and
// boundary arms, the day-portion cached iteration, calendar day-clamping, and
// the parseToSpan recover paths. All are pure/internal and need no Ruby.

// scalarTok builds a token carrying a scalar tag of the given class+value.
func scalarTok(class tagClass, n int) *token {
	t := newToken("x")
	t.tag(&tag{class: class, num: n})
	return t
}

// repTok builds a token carrying a repeater tag.
func repTok(class tagClass, r repeater) *token {
	t := newToken("x")
	t.tag(&tag{class: class, rep: r})
	return t
}

// pointerTok builds a pointer token.
func pointerTok(sym string) *token {
	t := newToken("x")
	t.tag(&tag{class: tcPointer, sym: sym})
	return t
}

func newBranchParser() *parser {
	return &parser{now: repeaterNow}
}

func TestHandleSRPNotOffsetter(t *testing.T) {
	withUTC(t)
	// A RepeaterSecond is an offsetter, but a RepeaterTime is NOT: handleSRP must
	// return nil when the repeater cannot offset.
	toks := []*token{
		scalarTok(tcScalar, 3),
		repTok(tcRepeaterTime, newRepeaterTime("5", &options{})),
		pointerTok("future"),
	}
	p := newBranchParser()
	if got := handleSRP(p, toks, &options{}); got != nil {
		t.Errorf("handleSRP with non-offsetter = %v, want nil", got)
	}
}

// bogusGrabberTok builds a grabber token with an unrecognised symbol, so
// getAnchor leaves its outer span nil (a clean, panic-free nil anchor).
func bogusGrabberTok() *token {
	tk := newToken("x")
	tk.tag(&tag{class: tcGrabber, sym: "bogus"})
	return tk
}

func TestHandleSRPADeadAnchor(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// tokens[3:] = {bogus-grabber, repeater} -> getAnchor returns nil -> handler nil.
	toks := []*token{
		scalarTok(tcScalar, 3),
		repTok(tcRepeaterDay, &repeaterDay{}),
		pointerTok("future"),
		bogusGrabberTok(),
		repTok(tcRepeaterDay, &repeaterDay{}),
	}
	if got := handleSRPA(p, toks, &options{}); got != nil {
		t.Errorf("handleSRPA(dead anchor) = %v, want nil", got)
	}
}

func TestHandleSRASRPADeadAnchor(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// tokens[4:] = {bogus-grabber, repeater, repeater} -> nil anchor -> nil.
	toks := []*token{
		scalarTok(tcScalar, 3), repTok(tcRepeaterDay, &repeaterDay{}),
		scalarTok(tcScalar, 2), repTok(tcRepeaterHour, &repeaterHour{}),
		bogusGrabberTok(), repTok(tcRepeaterDay, &repeaterDay{}), repTok(tcRepeaterDay, &repeaterDay{}),
	}
	if got := handleSRASRPA(p, toks, &options{}); got != nil {
		t.Errorf("handleSRASRPA(dead anchor) = %v, want nil", got)
	}
}

func TestHandleORGRDeadAnchor(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// ORGR anchor is tokens[2:4] = {bogus-grabber, repeater} -> nil.
	ord := newToken("x")
	ord.tag(&tag{class: tcOrdinal, num: 1})
	toks := []*token{ord, repTok(tcRepeaterDay, &repeaterDay{}), bogusGrabberTok(), repTok(tcRepeaterMonth, &repeaterMonth{})}
	if got := handleORGR(p, toks, &options{}); got != nil {
		t.Errorf("handleORGR(dead anchor) = %v, want nil", got)
	}
}

func TestGetAnchorNilOuter(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// A grabber sym that is none of last/this/next leaves outerSpan nil.
	grab := newToken("x")
	grab.tag(&tag{class: tcGrabber, sym: "bogus"})
	toks := []*token{grab, repTok(tcRepeaterDay, &repeaterDay{})}
	if got := p.getAnchor(toks, &options{context: "future"}); got != nil {
		t.Errorf("getAnchor(bogus grabber) = %v, want nil", got)
	}
}

func TestFindWithinNil(t *testing.T) {
	withUTC(t)
	// A repeater whose this() span lies entirely outside the outer span yields nil.
	// Outer window is a single second at midnight; a month repeater's "this" span
	// (the whole month, beginning on the 1st) is not covered by that instant.
	outer := newSpan(time.Date(2006, 8, 16, 12, 0, 0, 0, time.UTC), time.Date(2006, 8, 16, 12, 0, 1, 0, time.UTC))
	if got := findWithin([]repeater{&repeaterMonth{}}, outer, "future"); got != nil {
		t.Errorf("findWithin outside = %v, want nil", got)
	}
	// Empty tag list returns the span unchanged.
	if got := findWithin(nil, outer, "future"); got != outer {
		t.Error("findWithin(nil) should return the span")
	}
	// A repeater whose this() is covered by a wide outer span recurses and returns
	// the narrowed span (covers the future setStart + cover-recurse arms).
	wide := newSpan(time.Date(2006, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2006, 9, 1, 0, 0, 0, 0, time.UTC))
	if got := findWithin([]repeater{&repeaterDay{}}, wide, "future"); got == nil {
		t.Error("findWithin(day in month, future) = nil, want a span")
	}
	// The past pointer takes the setStart(span.End) arm.
	if got := findWithin([]repeater{&repeaterDay{}}, wide, "past"); got == nil {
		t.Error("findWithin(day in month, past) = nil, want a span")
	}
}

func TestHandleGenericBadTimestamp(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	if got := handleGeneric(p, nil, &options{text: "not-a-timestamp"}); got != nil {
		t.Errorf("handleGeneric(bad) = %v, want nil", got)
	}
}

func TestHandleRdnOdOverflow(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// day 31 in a future context that pushes into September (30 days) overflows.
	dn := repTok(tcRepeaterDayName, newRepeaterDayName("sunday"))
	od := scalarTok(tcOrdinalDay, 31)
	// Anchor now at Sep so month has 30 days; day 31 overflows -> nil.
	p.now = time.Date(2006, 9, 16, 14, 0, 0, 0, time.UTC)
	if got := handleRdnOd(p, []*token{dn, od}, &options{context: "future"}); got != nil {
		t.Errorf("handleRdnOd(31 in sept) = %v, want nil", got)
	}
}

func TestHandleRmnOdOnLongForm(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// len(tokens) > 3 branch: time, day-portion, month-name, ordinal-day.
	tm := repTok(tcRepeaterTime, newRepeaterTime("5", &options{}))
	dp := newToken("x")
	dp.tag(&tag{class: tcRepeaterDayPortion, sym: "pm", rep: newRepeaterDayPortion("pm", 0)})
	mn := repTok(tcRepeaterMonthName, newRepeaterMonthName("may"))
	od := scalarTok(tcOrdinalDay, 27)
	if got := handleRmnOdOn(p, []*token{tm, dp, mn, od}, &options{context: "future"}); got == nil {
		t.Error("handleRmnOdOn long form = nil, want a span")
	}
	// len(tokens) == 3 branch: time, month-name, ordinal-day (no day-portion).
	// Use a fresh parser: dayOrTime mutates p.now on the long-form call above.
	p2 := newBranchParser()
	tm2 := repTok(tcRepeaterTime, newRepeaterTime("5", &options{}))
	mn2 := repTok(tcRepeaterMonthName, newRepeaterMonthName("may"))
	od2 := scalarTok(tcOrdinalDay, 27)
	if got := handleRmnOdOn(p2, []*token{tm2, mn2, od2}, &options{context: "future"}); got == nil {
		t.Error("handleRmnOdOn short form = nil, want a span")
	}
}

func TestHandleSRASRPAMidNil(t *testing.T) {
	withUTC(t)
	p := newBranchParser()
	// Layout matches the SRASRPA grammar after separator filtering:
	// [scalar, rep, scalar, rep, pointer, anchor-repeater, anchor-repeater].
	// The FIRST inner arrow's repeater is a non-offsetter time, so its handleSRP
	// returns nil, exercising the mid-span nil guard.
	toks := []*token{
		scalarTok(tcScalar, 3), repTok(tcRepeaterTime, newRepeaterTime("5", &options{})),
		scalarTok(tcScalar, 2), repTok(tcRepeaterHour, &repeaterHour{}),
		pointerTok("future"), repTok(tcRepeaterDay, &repeaterDay{}), repTok(tcRepeaterMonth, &repeaterMonth{}),
	}
	if got := handleSRASRPA(p, toks, &options{context: "future"}); got != nil {
		t.Errorf("handleSRASRPA(mid nil) = %v, want nil", got)
	}
}

func TestMonthOverflowMonthBounds(t *testing.T) {
	if monthOverflow(2006, 0, 1) {
		t.Error("month 0 should not overflow (returns false)")
	}
	if monthOverflow(2006, 13, 1) {
		t.Error("month 13 should not overflow (returns false)")
	}
}

func TestNumerizeBigPrefixWithDigit(t *testing.T) {
	// "3 hundred" -> the big-prefix arm multiplies the leading digit.
	if got := numerize("3 hundred days"); got != "300 days" {
		t.Errorf("numerize(3 hundred days) = %q, want 300 days", got)
	}
	// bare "hundred" -> the empty-capture arm returns the prefix value.
	if got := numerize("hundred days"); got != "100 days" {
		t.Errorf("numerize(hundred days) = %q, want 100 days", got)
	}
}

func TestAnditionBoundaryAndOrder(t *testing.T) {
	// "<num>100 <num>5" with g1 longer than g3 additions.
	if got := andition("<num>100 <num>5"); got != "<num>105" {
		t.Errorf("andition(100 5) = %q, want <num>105", got)
	}
	// "and" separator forces addition even when g1 is not longer.
	if got := andition("<num>5 and <num>100"); got != "<num>105" {
		t.Errorf("andition(5 and 100) = %q, want <num>105", got)
	}
	// non-and separator with g1 not longer than g3 leaves the string unchanged.
	if got := andition("<num>5 <num>100"); got != "<num>5 <num>100" {
		t.Errorf("andition(5 100) = %q, want unchanged", got)
	}
	// a trailing word char after the match aborts (isWordByte guard).
	if got := andition("<num>100 <num>5x"); got != "<num>100 <num>5x" {
		t.Errorf("andition with word-suffix = %q, want unchanged", got)
	}
}

func TestParseToSpanRecoversOverflow(t *testing.T) {
	withUTC(t)
	// "next august" on an August anchor makes repeaterMonthName.next raise the
	// day-overflow sentinel, which parseToSpan recovers as a nil (no-parse) span.
	span := parseToSpan("next august", (&Options{Now: repeaterNow}).resolve())
	if span != nil {
		t.Errorf("parseToSpan(next august) = %v, want nil", span)
	}
}

func TestHandleParsePanic(t *testing.T) {
	// nil and the sentinel are swallowed.
	handleParsePanic(nil)
	handleParsePanic(errDayOverflow)
	// any other value re-panics.
	defer func() {
		if r := recover(); r != "foreign" {
			t.Fatalf("expected re-panic of foreign, got %v", r)
		}
	}()
	handleParsePanic("foreign")
}

func TestYearOffsetTimeLeapClamp(t *testing.T) {
	withUTC(t)
	// Feb 29 of a leap year offset to a non-leap year clamps to Feb 28.
	feb29 := time.Date(2004, 2, 29, 12, 0, 0, 0, time.UTC)
	got := yearOffsetTime(feb29, 1, 1) // 2005 is not a leap year
	if got.Month() != time.February || got.Day() != 28 {
		t.Errorf("yearOffsetTime feb29->2005 = %s, want feb 28", got)
	}
}

func TestMonthOffsetByLeapClamp(t *testing.T) {
	withUTC(t)
	// Jan 31 offset by one month into a non-leap February clamps to the 28th.
	got := monthOffsetBy(time.Date(2005, 1, 31, 0, 0, 0, 0, time.UTC), 1, "future")
	if got.Month() != time.February || got.Day() != 28 {
		t.Errorf("monthOffsetBy jan31 2005 = %s, want feb 28", got)
	}
	// Into a leap February the clamp lands on the 29th (leap-table arm).
	gotL := monthOffsetBy(time.Date(2004, 1, 31, 0, 0, 0, 0, time.UTC), 1, "future")
	if gotL.Month() != time.February || gotL.Day() != 29 {
		t.Errorf("monthOffsetBy jan31 2004 = %s, want feb 29", gotL)
	}
}

func TestWeekdayNextPastFirst(t *testing.T) {
	withUTC(t)
	r := &repeaterWeekday{}
	r.setStart(repeaterNow)
	if s := r.next("past"); s == nil || !s.Begin.Before(repeaterNow) {
		t.Errorf("weekday next(past) first = %v", s)
	}
	// From a Friday going forward, the next weekday skips Sat+Sun to Monday,
	// exercising the isWeekday skip-loop. 2006-08-18 is a Friday.
	rf := &repeaterWeekday{}
	rf.setStart(time.Date(2006, 8, 18, 12, 0, 0, 0, time.UTC))
	s := rf.next("future")
	if s.Begin.Weekday() != time.Monday {
		t.Errorf("weekday next(future) from Fri = %s, want Monday", s.Begin.Weekday())
	}
	// From a Monday going backward, the previous weekday skips Sun+Sat to Friday.
	rp := &repeaterWeekday{}
	rp.setStart(time.Date(2006, 8, 21, 12, 0, 0, 0, time.UTC)) // Monday
	sp := rp.next("past")
	if sp.Begin.Weekday() != time.Friday {
		t.Errorf("weekday next(past) from Mon = %s, want Friday", sp.Begin.Weekday())
	}
}

func TestDayPortionNextIterates(t *testing.T) {
	withUTC(t)
	r := newRepeaterDayPortion("morning", 0)
	r.setStart(repeaterNow)
	s1 := r.next("future")
	s2 := r.next("future") // cached-shift branch
	if s1 == nil || s2 == nil || !s2.Begin.After(s1.Begin) {
		t.Errorf("day-portion next did not advance: %v then %v", s1, s2)
	}
	// past shift.
	rp := newRepeaterDayPortion("evening", 0)
	rp.setStart(repeaterNow)
	p1 := rp.next("past")
	p2 := rp.next("past")
	if !p2.Begin.Before(p1.Begin) {
		t.Errorf("day-portion next(past) did not retreat: %v then %v", p1, p2)
	}
	// width after curSet uses the cached span width.
	if r.width() <= 0 {
		t.Error("day-portion width after set should be > 0")
	}
}

func TestRepeaterTimeUnambiguousWithMinutesSeconds(t *testing.T) {
	withUTC(t)
	// "13:05:07" exercises the minutes+seconds parse arms.
	r := newRepeaterTime("13:05:07", &options{})
	r.setStart(repeaterNow)
	s := r.next("future")
	if s.Begin.Hour() != 13 || s.Begin.Minute() != 5 || s.Begin.Second() != 7 {
		t.Errorf("13:05:07 -> %s", s.Begin)
	}
}

func TestRepeaterTimeSubseconds(t *testing.T) {
	withUTC(t)
	// A colon-separated 4-group time "13:05:07:250" splits into four parts, so the
	// fourth group drives the subseconds parse arm (250/10^3 = 0.25).
	r := newRepeaterTime("13:05:07:250", &options{})
	// 0.250 seconds = 250ms; the tick seconds carry the fractional part.
	frac := r.typ.seconds - float64(int(r.typ.seconds))
	if frac < 0.24 || frac > 0.26 {
		t.Errorf("subsecond fraction = %v, want ~0.25", frac)
	}
}

func TestRepeaterTimePastAmbiguous(t *testing.T) {
	withUTC(t)
	// An ambiguous single-digit hour resolved into the past picks the pm candidate.
	r := newRepeaterTime("5", &options{})
	r.setStart(repeaterNow)
	s := r.next("past")
	if s == nil || !s.Begin.Before(repeaterNow) {
		t.Errorf("ambiguous 5 past = %v", s)
	}
}

func TestFindCurrentSeasonEmpty(t *testing.T) {
	withUTC(t)
	r := &repeaterSeason{}
	// A miniDate that no season claims (e.g. the gap-free ring always claims one,
	// so we assert the winter wrap and a boundary). Dec 22 starts winter.
	if got := r.findCurrentSeason(miniDate{12, 22}); got != "winter" {
		t.Errorf("dec 22 = %q, want winter", got)
	}
}

func TestSeasonNameThisNone(t *testing.T) {
	withUTC(t)
	r := newRepeaterSeasonName("summer")
	r.setStart(repeaterNow)
	if s := r.this("none"); s == nil {
		t.Error("season-name this(none) = nil")
	}
}
