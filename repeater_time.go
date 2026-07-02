// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"math"
	"strings"
	"time"
)

// tick is Chronic::RepeaterTime::Tick — a seconds-since-midnight offset that
// also carries whether the source time was ambiguous (am/pm unknown).
type tick struct {
	seconds   float64
	ambiguous bool
}

// repeaterTime resolves a bare clock time (e.g. "5", "9:30", "13:45:30") to the
// next matching instant relative to @now, porting Chronic::RepeaterTime.
type repeaterTime struct {
	baseRepeater
	typ    tick
	cur    time.Time
	curSet bool
}

// newRepeaterTime parses the time word into a Tick, porting the initializer.
// A malformed word (>4 colon groups) yields a nil-returning constructor; the
// caller (scanTimes) has already gated on the time regexp, so this never
// receives >4 groups in practice, but we guard anyway.
func newRepeaterTime(word string, opts *options) *repeaterTime {
	parts := strings.Split(word, ":")
	if len(parts) > 4 {
		return &repeaterTime{typ: tick{}}
	}

	// Split a bare digit-run into h[:m[:s]] the way the gem does. Mirrors
	// RepeaterTime#initialize: it inserts the trailing two digits at index 1 and
	// truncates parts[0], re-reading parts[0] fresh each step (never a stale copy).
	if len(parts[0]) > 2 && len(parts) == 1 {
		if len(parts[0]) > 4 {
			secondIndex := len(parts[0]) - 2
			parts = insertAt(parts, 1, parts[0][secondIndex:])
			parts[0] = parts[0][:secondIndex]
		}
		minuteIndex := len(parts[0]) - 2
		parts = insertAt(parts, 1, parts[0][minuteIndex:])
		parts[0] = parts[0][:minuteIndex]
	}

	ambiguous := false
	hours := atoi(parts[0])

	hours24 := opts.hours24
	hours24Set := opts.hours24Set
	if !hours24Set || (hours24Set && !hours24) {
		if (len(parts[0]) == 1 && hours > 0) ||
			(hours >= 10 && hours <= 12) ||
			(hours24Set && !hours24 && hours > 0) {
			ambiguous = true
		}
		if hours == 12 && ambiguous {
			hours = 0
		}
	}

	hSec := hours * 60 * 60
	minutes := 0
	seconds := 0
	subseconds := 0.0
	if len(parts) > 1 {
		minutes = atoi(parts[1]) * 60
	}
	if len(parts) > 2 {
		seconds = atoi(parts[2])
	}
	if len(parts) > 3 {
		subseconds = float64(atoi(parts[3])) / math.Pow(10, float64(len(parts[3])))
	}
	return &repeaterTime{typ: tick{seconds: float64(hSec+minutes+seconds) + subseconds, ambiguous: ambiguous}}
}

func insertAt(s []string, i int, v string) []string {
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

func (r *repeaterTime) ambiguous() bool { return r.typ.ambiguous }

func (r *repeaterTime) next(pointer string) *Span {
	halfDay := 60 * 60 * 12
	fullDay := 60 * 60 * 24
	first := false

	if !r.curSet {
		first = true
		midnight := local(r.now.Year(), int(r.now.Month()), r.now.Day())
		yesterdayMidnight := midnight.Add(time.Duration(-fullDay) * time.Second)
		tomorrowMidnight := midnight.Add(time.Duration(fullDay) * time.Second)

		_, midOff := midnight.Zone()
		_, tomOff := tomorrowMidnight.Zone()
		offsetFix := time.Duration(midOff-tomOff) * time.Second
		tomorrowMidnight = tomorrowMidnight.Add(offsetFix)

		td := time.Duration(r.typ.seconds * float64(time.Second))
		var candidates []time.Time
		if pointer == "future" {
			if r.typ.ambiguous {
				candidates = []time.Time{
					midnight.Add(td).Add(offsetFix),
					midnight.Add(time.Duration(halfDay) * time.Second).Add(td).Add(offsetFix),
					tomorrowMidnight.Add(td),
				}
			} else {
				candidates = []time.Time{
					midnight.Add(td).Add(offsetFix),
					tomorrowMidnight.Add(td),
				}
			}
			for _, t := range candidates {
				if !t.Before(r.now) {
					r.cur = t
					break
				}
			}
		} else { // past
			if r.typ.ambiguous {
				candidates = []time.Time{
					midnight.Add(time.Duration(halfDay) * time.Second).Add(td).Add(offsetFix),
					midnight.Add(td).Add(offsetFix),
					yesterdayMidnight.Add(td).Add(time.Duration(halfDay) * time.Second),
				}
			} else {
				candidates = []time.Time{
					midnight.Add(td).Add(offsetFix),
					yesterdayMidnight.Add(td),
				}
			}
			for _, t := range candidates {
				if !t.After(r.now) {
					r.cur = t
					break
				}
			}
		}
		r.curSet = true
	}

	if !first {
		increment := fullDay
		if r.typ.ambiguous {
			increment = halfDay
		}
		if pointer == "future" {
			r.cur = r.cur.Add(time.Duration(increment) * time.Second)
		} else {
			r.cur = r.cur.Add(time.Duration(-increment) * time.Second)
		}
	}
	return newSpan(r.cur, r.cur.Add(time.Duration(r.width())*time.Second))
}

func (r *repeaterTime) this(pointer string) *Span {
	if pointer == "none" {
		pointer = "future"
	}
	return r.next(pointer)
}

func (r *repeaterTime) width() int { return 1 }
