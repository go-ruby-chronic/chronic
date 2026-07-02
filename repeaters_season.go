// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// seasons maps a season name to its MiniDate span, mirroring
// RepeaterSeason::SEASONS.
var seasons = map[string]season{
	"spring": {miniDate{3, 20}, miniDate{6, 20}},
	"summer": {miniDate{6, 21}, miniDate{9, 22}},
	"autumn": {miniDate{9, 23}, miniDate{12, 21}},
	"winter": {miniDate{12, 22}, miniDate{3, 19}},
}

// --- RepeaterSeason ---

type repeaterSeason struct {
	baseRepeater
	nextStart time.Time
	nextEnd   time.Time
	seeded    bool
}

func (r *repeaterSeason) next(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	nextSeason := findNextSeason(r.findCurrentSeason(miniDateFromTime(r.now)), dir)
	return r.findNextSeasonSpan(dir, nextSeason)
}

func (r *repeaterSeason) this(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	today := construct(r.now.Year(), int(r.now.Month()), r.now.Day())
	thisSsn := r.findCurrentSeason(miniDateFromTime(r.now))
	var start, end time.Time
	switch pointer {
	case "past":
		start = today.Add(time.Duration(dir*r.numSecondsTilStart(thisSsn, dir)) * time.Second)
		end = today
	case "future":
		start = today.Add(daySeconds * time.Second)
		end = today.Add(time.Duration(dir*r.numSecondsTilEnd(thisSsn, dir)) * time.Second)
	default: // none
		start = today.Add(time.Duration(dir*r.numSecondsTilStart(thisSsn, dir)) * time.Second)
		end = today.Add(time.Duration(dir*r.numSecondsTilEnd(thisSsn, dir)) * time.Second)
	}
	return r.constructSeason(start, end)
}

func (r *repeaterSeason) offset(span *Span, amount int, pointer string) *Span {
	return newSpan(r.offsetBy(span.Begin, amount, pointer), r.offsetBy(span.End, amount, pointer))
}

func (r *repeaterSeason) offsetBy(t time.Time, amount int, pointer string) time.Time {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return t.Add(time.Duration(amount*dir*seasonSeconds) * time.Second)
}

func (r *repeaterSeason) width() int { return seasonSeconds }

func (r *repeaterSeason) findNextSeasonSpan(dir int, nextSeason string) *Span {
	if !r.seeded {
		r.nextStart = construct(r.now.Year(), int(r.now.Month()), r.now.Day())
		r.nextEnd = construct(r.now.Year(), int(r.now.Month()), r.now.Day())
		r.seeded = true
	}
	r.nextStart = r.nextStart.Add(time.Duration(dir*r.numSecondsTilStart(nextSeason, dir)) * time.Second)
	r.nextEnd = r.nextEnd.Add(time.Duration(dir*r.numSecondsTilEnd(nextSeason, dir)) * time.Second)
	return r.constructSeason(r.nextStart, r.nextEnd)
}

func (r *repeaterSeason) findCurrentSeason(md miniDate) string {
	// The four season ranges tile the year with no gaps, so any date not in the
	// first three must be winter; returning the last season as the default keeps
	// the logic total without an unreachable fall-through branch.
	for _, s := range seasonOrder[:len(seasonOrder)-1] {
		if md.isBetween(seasons[s].start, seasons[s].end) {
			return s
		}
	}
	return seasonOrder[len(seasonOrder)-1]
}

func (r *repeaterSeason) numSecondsTil(goal miniDate, dir int) int {
	start := construct(r.now.Year(), int(r.now.Month()), r.now.Day())
	seconds := 0
	for !miniDateFromTime(start.Add(time.Duration(dir*seconds) * time.Second)).equals(goal) {
		seconds += daySeconds
	}
	return seconds
}

func (r *repeaterSeason) numSecondsTilStart(s string, dir int) int {
	return r.numSecondsTil(seasons[s].start, dir)
}

func (r *repeaterSeason) numSecondsTilEnd(s string, dir int) int {
	return r.numSecondsTil(seasons[s].end, dir)
}

func (r *repeaterSeason) constructSeason(start, finish time.Time) *Span {
	return newSpan(
		construct(start.Year(), int(start.Month()), start.Day()),
		construct(finish.Year(), int(finish.Month()), finish.Day()),
	)
}

// --- RepeaterSeasonName ---

type repeaterSeasonName struct {
	repeaterSeason
	name string
}

func newRepeaterSeasonName(name string) *repeaterSeasonName {
	return &repeaterSeasonName{name: name}
}

func (r *repeaterSeasonName) next(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return r.findNextSeasonSpan(dir, r.name)
}

func (r *repeaterSeasonName) this(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	today := construct(r.now.Year(), int(r.now.Month()), r.now.Day())
	goalStart := today.Add(time.Duration(dir*r.numSecondsTilStart(r.name, dir)) * time.Second)
	goalEnd := today.Add(time.Duration(dir*r.numSecondsTilEnd(r.name, dir)) * time.Second)
	curr := r.findCurrentSeason(miniDateFromTime(r.now))
	var start, end time.Time
	switch pointer {
	case "past":
		start = goalStart
		if curr == r.name {
			end = today
		} else {
			end = goalEnd
		}
	case "future":
		if curr == r.name {
			start = today.Add(daySeconds * time.Second)
		} else {
			start = goalStart
		}
		end = goalEnd
	default:
		start = goalStart
		end = goalEnd
	}
	return r.constructSeason(start, end)
}

func (r *repeaterSeasonName) offset(span *Span, amount int, pointer string) *Span {
	return newSpan(r.offsetByName(span.Begin, amount, pointer), r.offsetByName(span.End, amount, pointer))
}

func (r *repeaterSeasonName) offsetByName(t time.Time, amount int, pointer string) time.Time {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return t.Add(time.Duration(amount*dir*yearSeconds) * time.Second)
}
