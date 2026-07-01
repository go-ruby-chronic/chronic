// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "strconv"

// This file ports chronic 0.10.2's Chronic::Date and Chronic::Time numeric
// predicates (no width parameter in 0.10.2) plus two-digit-year resolution.

func couldBeDay(day int) bool { return day >= 1 && day <= 31 }

func couldBeMonth(month int) bool { return month >= 1 && month <= 12 }

func couldBeYear(year int) bool { return year >= 1 && year <= 9999 }

func couldBeHour(hour int, hours12 bool) bool {
	max := 24
	if hours12 {
		max = 12
	}
	return hour >= 0 && hour <= max
}

func couldBeMinute(minute int) bool { return minute >= 0 && minute <= 60 }

func couldBeSecond(second int) bool { return second >= 0 && second <= 60 }

func couldBeSubsecond(sub int) bool { return sub >= 0 && sub <= 999999 }

// makeYear expands a two-digit year to four digits using the ambiguous future
// bias, mirroring Chronic::Date.make_year. nowYear is the anchor year.
func makeYear(year, bias, nowYear int) int {
	if len(strconv.Itoa(year)) > 2 {
		return year
	}
	startYear := nowYear - bias
	century := (startYear / 100) * 100
	fullYear := century + year
	if fullYear < startYear {
		fullYear += 100
	}
	return fullYear
}
