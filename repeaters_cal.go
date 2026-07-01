// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// --- RepeaterYear ---

type repeaterYear struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterYear) next(pointer string) *Span {
	if !r.curSet {
		if pointer == "future" {
			r.cur = construct(r.now.Year() + 1)
		} else {
			r.cur = construct(r.now.Year() - 1)
		}
		r.curSet = true
	} else {
		diff := 1
		if pointer == "past" {
			diff = -1
		}
		r.cur = construct(r.cur.Year() + diff)
	}
	return newSpan(r.cur, construct(r.cur.Year()+1))
}

func (r *repeaterYear) this(pointer string) *Span {
	switch pointer {
	case "future":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day()+1), construct(r.now.Year()+1, 1, 1))
	case "past":
		return newSpan(construct(r.now.Year(), 1, 1), construct(r.now.Year(), int(r.now.Month()), r.now.Day()))
	default:
		return newSpan(construct(r.now.Year(), 1, 1), construct(r.now.Year()+1, 1, 1))
	}
}

func (r *repeaterYear) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return newSpan(yearOffsetTime(span.Begin, amount, dir), yearOffsetTime(span.End, amount, dir))
}

func yearOffsetTime(t time.Time, amount, dir int) time.Time {
	year := t.Year() + amount*dir
	table := monthDays
	if isLeap(year) {
		table = monthDaysLeap
	}
	days := table[int(t.Month())]
	day := t.Day()
	if day > days {
		day = days
	}
	return construct(year, int(t.Month()), day, t.Hour(), t.Minute(), t.Second())
}

func (r *repeaterYear) width() int { return yearSeconds }

// --- RepeaterMonth ---

type repeaterMonth struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterMonth) next(pointer string) *Span {
	if !r.curSet {
		r.cur = monthOffsetBy(construct(r.now.Year(), int(r.now.Month())), 1, pointer)
		r.curSet = true
	} else {
		r.cur = monthOffsetBy(construct(r.cur.Year(), int(r.cur.Month())), 1, pointer)
	}
	return newSpan(r.cur, construct(r.cur.Year(), int(r.cur.Month())+1))
}

func (r *repeaterMonth) this(pointer string) *Span {
	switch pointer {
	case "future":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day()+1),
			monthOffsetBy(construct(r.now.Year(), int(r.now.Month())), 1, "future"))
	case "past":
		return newSpan(construct(r.now.Year(), int(r.now.Month())),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day()))
	default:
		return newSpan(construct(r.now.Year(), int(r.now.Month())),
			monthOffsetBy(construct(r.now.Year(), int(r.now.Month())), 1, "future"))
	}
}

func (r *repeaterMonth) offset(span *Span, amount int, pointer string) *Span {
	return newSpan(monthOffsetBy(span.Begin, amount, pointer), monthOffsetBy(span.End, amount, pointer))
}

func monthOffsetBy(t time.Time, amount int, pointer string) time.Time {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	amountYears := dir * amount / yearMonths
	amountMonths := dir * amount % yearMonths
	newYear := t.Year() + amountYears
	newMonth := int(t.Month()) + amountMonths
	if newMonth > yearMonths {
		newYear++
		newMonth -= yearMonths
	}
	table := monthDays
	if isLeap(newYear) {
		table = monthDaysLeap
	}
	days := table[newMonth]
	newDay := t.Day()
	if newDay > days {
		newDay = days
	}
	return construct(newYear, newMonth, newDay, t.Hour(), t.Minute(), t.Second())
}

func (r *repeaterMonth) width() int { return 30 * daySeconds }

// --- RepeaterWeek ---

type repeaterWeek struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterWeek) next(pointer string) *Span {
	if !r.curSet {
		if pointer == "future" {
			dn := newRepeaterDayName("sunday")
			dn.setStart(r.now)
			r.cur = dn.next("future").Begin
		} else {
			dn := newRepeaterDayName("sunday")
			dn.setStart(r.now.Add(daySeconds * time.Second))
			dn.next("past")
			r.cur = dn.next("past").Begin
		}
		r.curSet = true
	} else {
		dir := 1
		if pointer == "past" {
			dir = -1
		}
		r.cur = r.cur.Add(time.Duration(dir*weekSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(weekSeconds*time.Second))
}

func (r *repeaterWeek) this(pointer string) *Span {
	switch pointer {
	case "future":
		start := local(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()).Add(hourSeconds * time.Second)
		dn := newRepeaterDayName("sunday")
		dn.setStart(r.now)
		end := dn.this("future").Begin
		return newSpan(start, end)
	case "past":
		end := local(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour())
		dn := newRepeaterDayName("sunday")
		dn.setStart(r.now)
		start := dn.next("past").Begin
		return newSpan(start, end)
	default:
		dn := newRepeaterDayName("sunday")
		dn.setStart(r.now)
		start := dn.next("past").Begin
		return newSpan(start, start.Add(weekSeconds*time.Second))
	}
}

func (r *repeaterWeek) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * weekSeconds)
}

func (r *repeaterWeek) width() int { return weekSeconds }

// --- RepeaterWeekday ---

var weekdayDayNums = map[string]int{
	"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3,
	"thursday": 4, "friday": 5, "saturday": 6,
}

type repeaterWeekday struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func isWeekday(wday int) bool {
	return wday != 0 && wday != 6
}

func (r *repeaterWeekday) next(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	if !r.curSet {
		r.cur = construct(r.now.Year(), int(r.now.Month()), r.now.Day())
		r.cur = r.cur.Add(time.Duration(dir*daySeconds) * time.Second)
		for !isWeekday(int(r.cur.Weekday())) {
			r.cur = r.cur.Add(time.Duration(dir*daySeconds) * time.Second)
		}
		r.curSet = true
	} else {
		for {
			r.cur = r.cur.Add(time.Duration(dir*daySeconds) * time.Second)
			if isWeekday(int(r.cur.Weekday())) {
				break
			}
		}
	}
	return newSpan(r.cur, r.cur.Add(daySeconds*time.Second))
}

func (r *repeaterWeekday) this(pointer string) *Span {
	if pointer == "past" {
		return r.next("past")
	}
	return r.next("future")
}

func (r *repeaterWeekday) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	passed := 0
	off := 0
	for passed != amount {
		off += dir * daySeconds
		if isWeekday(int(span.Begin.Add(time.Duration(off) * time.Second).Weekday())) {
			passed++
		}
	}
	return span.addSeconds(off)
}

func (r *repeaterWeekday) width() int { return daySeconds }

// --- RepeaterWeekend ---

type repeaterWeekend struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterWeekend) next(pointer string) *Span {
	if !r.curSet {
		sat := newRepeaterDayName("saturday")
		if pointer == "future" {
			sat.setStart(r.now)
			r.cur = sat.next("future").Begin
		} else {
			sat.setStart(r.now.Add(daySeconds * time.Second))
			r.cur = sat.next("past").Begin
		}
		r.curSet = true
	} else {
		dir := 1
		if pointer == "past" {
			dir = -1
		}
		r.cur = r.cur.Add(time.Duration(dir*weekSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(weekendSeconds*time.Second))
}

func (r *repeaterWeekend) this(pointer string) *Span {
	sat := newRepeaterDayName("saturday")
	sat.setStart(r.now)
	if pointer == "past" {
		b := sat.this("past").Begin
		return newSpan(b, b.Add(weekendSeconds*time.Second))
	}
	b := sat.this("future").Begin
	return newSpan(b, b.Add(weekendSeconds*time.Second))
}

func (r *repeaterWeekend) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	w := &repeaterWeekend{}
	w.setStart(span.Begin)
	start := w.next(pointer).Begin.Add(time.Duration((amount-1)*dir*weekSeconds) * time.Second)
	return newSpan(start, start.Add(span.End.Sub(span.Begin)))
}

func (r *repeaterWeekend) width() int { return weekendSeconds }

// --- RepeaterFortnight ---

type repeaterFortnight struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterFortnight) next(pointer string) *Span {
	if !r.curSet {
		sun := newRepeaterDayName("sunday")
		if pointer == "future" {
			sun.setStart(r.now)
			r.cur = sun.next("future").Begin
		} else {
			sun.setStart(r.now.Add(daySeconds * time.Second))
			sun.next("past")
			sun.next("past")
			r.cur = sun.next("past").Begin
		}
		r.curSet = true
	} else {
		dir := 1
		if pointer == "past" {
			dir = -1
		}
		r.cur = r.cur.Add(time.Duration(dir*fortnightSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(fortnightSeconds*time.Second))
}

func (r *repeaterFortnight) this(pointer string) *Span {
	if pointer == "none" {
		pointer = "future"
	}
	if pointer == "past" {
		end := construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour())
		sun := newRepeaterDayName("sunday")
		sun.setStart(r.now)
		start := sun.next("past").Begin
		return newSpan(start, end)
	}
	start := construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()).Add(hourSeconds * time.Second)
	sun := newRepeaterDayName("sunday")
	sun.setStart(r.now)
	sun.this("future")
	end := sun.this("future").Begin
	return newSpan(start, end)
}

func (r *repeaterFortnight) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * fortnightSeconds)
}

func (r *repeaterFortnight) width() int { return fortnightSeconds }
