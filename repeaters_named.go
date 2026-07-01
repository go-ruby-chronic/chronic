// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// --- RepeaterDayName ---

type repeaterDayName struct {
	baseRepeater
	dayName string
	cur     time.Time
	curSet  bool
}

func newRepeaterDayName(name string) *repeaterDayName {
	return &repeaterDayName{dayName: name}
}

func (r *repeaterDayName) next(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	if !r.curSet {
		r.cur = dateOnly(r.now)
		r.cur = r.cur.AddDate(0, 0, dir)
		dayNum := weekdayDayNums[r.dayName]
		for int(r.cur.Weekday()) != dayNum {
			r.cur = r.cur.AddDate(0, 0, dir)
		}
		r.curSet = true
	} else {
		r.cur = r.cur.AddDate(0, 0, dir*7)
	}
	nextDate := r.cur.AddDate(0, 0, 1)
	return newSpan(
		construct(r.cur.Year(), int(r.cur.Month()), r.cur.Day()),
		construct(nextDate.Year(), int(nextDate.Month()), nextDate.Day()),
	)
}

func (r *repeaterDayName) this(pointer string) *Span {
	if pointer == "none" {
		pointer = "future"
	}
	return r.next(pointer)
}

func (r *repeaterDayName) width() int { return daySeconds }

// dateOnly returns midnight of t in TimeLoc, mirroring Date.new(y,m,d).
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, TimeLoc)
}

// --- RepeaterMonthName ---

var monthNameIndex = map[string]int{
	"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6,
	"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
}

type repeaterMonthName struct {
	baseRepeater
	name   string
	cur    time.Time
	curSet bool
}

func newRepeaterMonthName(name string) *repeaterMonthName {
	return &repeaterMonthName{name: name}
}

func (r *repeaterMonthName) index() int { return monthNameIndex[r.name] }

func (r *repeaterMonthName) next(pointer string) *Span {
	idx := r.index()
	if !r.curSet {
		assigned := true
		switch pointer {
		case "future":
			if int(r.now.Month()) < idx {
				r.cur = construct(r.now.Year(), idx)
			} else if int(r.now.Month()) > idx {
				r.cur = construct(r.now.Year()+1, idx)
			} else {
				assigned = false
			}
		case "none":
			if int(r.now.Month()) <= idx {
				r.cur = construct(r.now.Year(), idx)
			} else {
				r.cur = construct(r.now.Year()+1, idx)
			}
		case "past":
			if int(r.now.Month()) >= idx {
				r.cur = construct(r.now.Year(), idx)
			} else {
				r.cur = construct(r.now.Year()-1, idx)
			}
		}
		// The gem raises 'Current month should be set by now' when @now.month
		// equals the target month with a :future/:past pointer; that exception
		// propagates out of Chronic.parse. We surface it as a no-parse (nil) via
		// the recovered sentinel, so "next august" on an August anchor yields nil
		// rather than a bogus date.
		if !assigned {
			panic(errDayOverflow)
		}
		r.curSet = true
	} else {
		if pointer == "future" {
			r.cur = construct(r.cur.Year()+1, int(r.cur.Month()))
		} else if pointer == "past" {
			r.cur = construct(r.cur.Year()-1, int(r.cur.Month()))
		}
	}

	y, m := r.cur.Year(), int(r.cur.Month())
	var ny, nm int
	if m == 12 {
		ny, nm = y+1, 1
	} else {
		ny, nm = y, m+1
	}
	return newSpan(r.cur, construct(ny, nm))
}

func (r *repeaterMonthName) this(pointer string) *Span {
	if pointer == "past" {
		return r.next("past")
	}
	// future or none
	return r.next("none")
}

func (r *repeaterMonthName) width() int { return 30 * daySeconds }

// --- RepeaterDayPortion ---

type dpRange struct{ begin, end int }

var dpPortions = map[string]dpRange{
	"am":        {0, 12*3600 - 1},
	"pm":        {12 * 3600, 24*3600 - 1},
	"morning":   {6 * 3600, 12 * 3600},
	"afternoon": {13 * 3600, 17 * 3600},
	"evening":   {17 * 3600, 20 * 3600},
	"night":     {20 * 3600, 24 * 3600},
}

type repeaterDayPortion struct {
	baseRepeater
	sym     string // named portion, empty when integer
	intType int    // integer hour boundary (ambiguous_time_range)
	isInt   bool
	rng     dpRange
	cur     *Span
	curSet  bool
}

func newRepeaterDayPortion(sym string, intType int) *repeaterDayPortion {
	r := &repeaterDayPortion{sym: sym, intType: intType, isInt: sym == ""}
	if r.isInt {
		r.rng = dpRange{intType * 3600, (intType + 12) * 3600}
	} else {
		r.rng = dpPortions[sym]
	}
	return r
}

func (r *repeaterDayPortion) next(pointer string) *Span {
	if !r.curSet {
		nowSeconds := int(r.now.Sub(construct(r.now.Year(), int(r.now.Month()), r.now.Day())).Seconds())
		var rangeStart time.Time
		base := construct(r.now.Year(), int(r.now.Month()), r.now.Day())
		if nowSeconds < r.rng.begin {
			if pointer == "future" {
				rangeStart = base.Add(time.Duration(r.rng.begin) * time.Second)
			} else {
				rangeStart = construct(r.now.Year(), int(r.now.Month()), r.now.Day()-1).Add(time.Duration(r.rng.begin) * time.Second)
			}
		} else if nowSeconds > r.rng.end {
			if pointer == "future" {
				rangeStart = construct(r.now.Year(), int(r.now.Month()), r.now.Day()+1).Add(time.Duration(r.rng.begin) * time.Second)
			} else {
				rangeStart = base.Add(time.Duration(r.rng.begin) * time.Second)
			}
		} else {
			if pointer == "future" {
				rangeStart = construct(r.now.Year(), int(r.now.Month()), r.now.Day()+1).Add(time.Duration(r.rng.begin) * time.Second)
			} else {
				rangeStart = construct(r.now.Year(), int(r.now.Month()), r.now.Day()-1).Add(time.Duration(r.rng.begin) * time.Second)
			}
		}
		offset := r.rng.end - r.rng.begin
		rangeEnd := r.dateFromRefOffset(rangeStart, offset)
		r.cur = newSpan(rangeStart, rangeEnd)
		r.curSet = true
	} else {
		shift := 1
		if pointer == "past" {
			shift = -1
		}
		nb := construct(r.cur.Begin.Year(), int(r.cur.Begin.Month()), r.cur.Begin.Day()+shift, r.cur.Begin.Hour(), r.cur.Begin.Minute(), r.cur.Begin.Second())
		ne := construct(r.cur.End.Year(), int(r.cur.End.Month()), r.cur.End.Day()+shift, r.cur.End.Hour(), r.cur.End.Minute(), r.cur.End.Second())
		r.cur = newSpan(nb, ne)
	}
	return r.cur
}

func (r *repeaterDayPortion) this(pointer string) *Span {
	rangeStart := construct(r.now.Year(), int(r.now.Month()), r.now.Day()).Add(time.Duration(r.rng.begin) * time.Second)
	rangeEnd := r.dateFromRefOffset(rangeStart, -1)
	r.cur = newSpan(rangeStart, rangeEnd)
	r.curSet = true
	return r.cur
}

func (r *repeaterDayPortion) offset(span *Span, amount int, pointer string) *Span {
	r.now = span.Begin
	portionSpan := r.next(pointer)
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return portionSpan.addSeconds(dir * (amount - 1) * daySeconds)
}

func (r *repeaterDayPortion) width() int {
	if r.curSet {
		return r.cur.Width()
	}
	if r.isInt {
		return 12 * 3600
	}
	return r.rng.end - r.rng.begin
}

// dateFromRefOffset ports construct_date_from_reference_and_offset. offset < 0
// means "use the range width" (the Ruby default-nil).
func (r *repeaterDayPortion) dateFromRefOffset(reference time.Time, offset int) time.Time {
	elapsed := offset
	if offset < 0 {
		elapsed = r.rng.end - r.rng.begin
	}
	secondHand := ((elapsed - (12 * 60)) % 60)
	secondHand = ((secondHand % 60) + 60) % 60
	minuteHand := ((elapsed - secondHand) / 60) % 60
	hourHand := (elapsed-minuteHand-secondHand)/(60*60) + reference.Hour()%24
	return construct(reference.Year(), int(reference.Month()), reference.Day(), hourHand, minuteHand, secondHand)
}
