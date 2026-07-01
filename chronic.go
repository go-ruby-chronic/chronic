// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package chronic is a pure-Go (no cgo) reimplementation of Ruby's `chronic`
// gem — a natural-language date and time parser. It turns English phrases such
// as "tomorrow", "next tuesday", "3 days ago", "may 27th" or "2016-05-27" into
// a concrete instant (or a [Span]), faithfully mirroring the deterministic
// grammar of chronic 0.10.2.
//
// The entry point is [Parse]. It is deterministic: pass a fixed anchor via
// [Options.Now] and the result depends only on the input string, so it needs no
// Ruby runtime.
//
//	now := time.Date(2006, 8, 16, 14, 0, 0, 0, time.Local)
//	t, _ := chronic.Parse("tomorrow", chronic.Options{Now: now})
//	// t == 2006-08-17 12:00:00
package chronic

import "time"

// TimeLoc is the location every constructed instant lives in. It mirrors MRI's
// Time.local, which builds times in the process-local zone. It is a package
// variable so callers (and the oracle tests) can pin it to a fixed zone for
// reproducibility.
var TimeLoc = time.Local

// Handler support constants shared across the grammar (seconds unless noted).
const (
	yearSeconds      = 31_536_000 // 365 days
	seasonSeconds    = 7_862_400  // 91 days
	quarterSeconds   = 7_776_000  // 90 days
	fortnightSeconds = 1_209_600  // 14 days
	weekSeconds      = 604_800    // 7 days
	weekendSeconds   = 172_800    // 2 days
	daySeconds       = 86_400
	hourSeconds      = 3600
	minuteSeconds    = 60
	secondSeconds    = 1
	yearMonths       = 12
)

// monthDays / monthDaysLeap are indexed 1..12 (index 0 unused) to match the
// gem's Chronic::Date tables used by construct's overflow logic.
var (
	monthDays     = []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	monthDaysLeap = []int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
)

// isLeap reports whether year is a Gregorian leap year, matching Ruby Date.leap?.
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// construct builds a time from the given components, resolving second/minute/
// hour/day/month overflow exactly as Chronic.construct does. It panics with a
// value recovered by the parser when day exceeds 56 (the gem raises there).
func construct(year int, rest ...int) time.Time {
	month, day, hour, minute, second := 1, 1, 0, 0, 0
	if len(rest) > 0 {
		month = rest[0]
	}
	if len(rest) > 1 {
		day = rest[1]
	}
	if len(rest) > 2 {
		hour = rest[2]
	}
	if len(rest) > 3 {
		minute = rest[3]
	}
	if len(rest) > 4 {
		second = rest[4]
	}

	if second >= 60 {
		minute += second / 60
		second = second % 60
	}
	if minute >= 60 {
		hour += minute / 60
		minute = minute % 60
	}
	if hour >= 24 {
		day += hour / 24
		hour = hour % 24
	}

	if day > 56 {
		panic(errDayOverflow)
	}
	if day > 28 {
		table := monthDays
		if isLeap(year) {
			table = monthDaysLeap
		}
		daysThisMonth := table[month]
		if day > daysThisMonth {
			month += day / daysThisMonth
			day = day % daysThisMonth
		}
	}

	if month > 12 {
		if month%12 == 0 {
			year += (month - 12) / 12
			month = 12
		} else {
			year += month / 12
			month = month % 12
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, TimeLoc)
}

// errDayOverflow is panicked by construct when the day count is unresolvable.
var errDayOverflow = "day must be no more than 56 (makes month resolution easier)"

// local mirrors Time.local(year, month=1, day=1, hour=0, min=0, sec=0). Unlike
// construct it does not resolve overflow: Go's time.Date normalises out-of-range
// components, which matches how MRI raises/normalises for these callers combined
// with the surrounding month_overflow? guards.
func local(year int, rest ...int) time.Time {
	month, day, hour, minute, second := 1, 1, 0, 0, 0
	switch len(rest) {
	case 5:
		second = rest[4]
		fallthrough
	case 4:
		minute = rest[3]
		fallthrough
	case 3:
		hour = rest[2]
		fallthrough
	case 2:
		day = rest[1]
		fallthrough
	case 1:
		month = rest[0]
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, TimeLoc)
}
