// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// miniDate is a month/day pair used to classify seasons, mirroring
// Chronic::MiniDate.
type miniDate struct {
	month int
	day   int
}

func miniDateFromTime(t time.Time) miniDate {
	return miniDate{month: int(t.Month()), day: t.Day()}
}

// isBetween reports whether the receiver falls within [start, end] treating the
// calendar as a wrapping ring of months, mirroring MiniDate#is_between?.
func (m miniDate) isBetween(start, end miniDate) bool {
	if m.month == start.month && m.month == end.month &&
		(m.day < start.day || m.day > end.day) {
		return false
	}
	if (m.month == start.month && m.day >= start.day) ||
		(m.month == end.month && m.day <= end.day) {
		return true
	}
	i := (start.month % 12) + 1
	for i != end.month {
		if m.month == i {
			return true
		}
		i = (i % 12) + 1
	}
	return false
}

func (m miniDate) equals(other miniDate) bool {
	return m.month == other.month && m.day == other.day
}

// season is a span of MiniDates, mirroring Chronic::Season.
type season struct {
	start miniDate
	end   miniDate
}

var seasonOrder = []string{"spring", "summer", "autumn", "winter"}

// findNextSeason returns the season pointer steps away (wrapping), mirroring
// Season.find_next_season.
func findNextSeason(s string, pointer int) string {
	idx := indexOf(seasonOrder, s)
	next := ((idx + 1*pointer) % 4 + 4) % 4
	return seasonOrder[next]
}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}
