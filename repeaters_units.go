// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// --- RepeaterDay ---

type repeaterDay struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterDay) next(pointer string) *Span {
	if !r.curSet {
		r.cur = local(r.now.Year(), int(r.now.Month()), r.now.Day())
		r.curSet = true
	}
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	r.cur = r.cur.Add(time.Duration(dir*daySeconds) * time.Second)
	return newSpan(r.cur, r.cur.Add(daySeconds*time.Second))
}

func (r *repeaterDay) this(pointer string) *Span {
	switch pointer {
	case "future":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day()).Add(daySeconds*time.Second))
	case "past":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day()),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()))
	default: // none
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day()),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day()).Add(daySeconds*time.Second))
	}
}

func (r *repeaterDay) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * daySeconds)
}

func (r *repeaterDay) width() int { return daySeconds }

// --- RepeaterHour ---

type repeaterHour struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterHour) next(pointer string) *Span {
	if !r.curSet {
		if pointer == "future" {
			r.cur = construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()+1)
		} else {
			r.cur = construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()-1)
		}
		r.curSet = true
	} else {
		dir := 1
		if pointer == "past" {
			dir = -1
		}
		r.cur = r.cur.Add(time.Duration(dir*hourSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(hourSeconds*time.Second))
}

func (r *repeaterHour) this(pointer string) *Span {
	switch pointer {
	case "future":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()+1),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()+1))
	case "past":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour()),
			construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()))
	default:
		hs := construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour())
		return newSpan(hs, hs.Add(hourSeconds*time.Second))
	}
}

func (r *repeaterHour) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * hourSeconds)
}

func (r *repeaterHour) width() int { return hourSeconds }

// --- RepeaterMinute ---

type repeaterMinute struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterMinute) next(pointer string) *Span {
	if !r.curSet {
		if pointer == "future" {
			r.cur = construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()+1)
		} else {
			r.cur = construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()-1)
		}
		r.curSet = true
	} else {
		dir := 1
		if pointer == "past" {
			dir = -1
		}
		r.cur = r.cur.Add(time.Duration(dir*minuteSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(minuteSeconds*time.Second))
}

func (r *repeaterMinute) this(pointer string) *Span {
	switch pointer {
	case "future":
		return newSpan(r.now, construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()))
	case "past":
		return newSpan(construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute()), r.now)
	default:
		mb := construct(r.now.Year(), int(r.now.Month()), r.now.Day(), r.now.Hour(), r.now.Minute())
		return newSpan(mb, mb.Add(minuteSeconds*time.Second))
	}
}

func (r *repeaterMinute) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * minuteSeconds)
}

func (r *repeaterMinute) width() int { return minuteSeconds }

// --- RepeaterSecond ---

type repeaterSecond struct {
	baseRepeater
	cur    time.Time
	curSet bool
}

func (r *repeaterSecond) next(pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	if !r.curSet {
		r.cur = r.now.Add(time.Duration(dir*secondSeconds) * time.Second)
		r.curSet = true
	} else {
		r.cur = r.cur.Add(time.Duration(dir*secondSeconds) * time.Second)
	}
	return newSpan(r.cur, r.cur.Add(secondSeconds*time.Second))
}

func (r *repeaterSecond) this(pointer string) *Span {
	return newSpan(r.now, r.now.Add(time.Second))
}

func (r *repeaterSecond) offset(span *Span, amount int, pointer string) *Span {
	dir := 1
	if pointer == "past" {
		dir = -1
	}
	return span.addSeconds(dir * amount * secondSeconds)
}

func (r *repeaterSecond) width() int { return secondSeconds }
