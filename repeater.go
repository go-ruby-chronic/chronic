// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"regexp"
	"time"
)

// repeater is the behaviour a Chronic::Repeater exposes to the handler pipeline.
// Each concrete repeater keeps its own iteration state and a start instant
// (@now in the gem), set via setStart.
type repeater interface {
	setStart(t time.Time)
	next(pointer string) *Span
	this(pointer string) *Span
	width() int
}

// offsetter is implemented by repeaters that can shift a span by a scalar
// amount. Mirrors `respond_to?(:offset)`.
type offsetter interface {
	offset(span *Span, amount int, pointer string) *Span
}

// baseRepeater holds the shared @now field.
type baseRepeater struct {
	now time.Time
}

func (b *baseRepeater) setStart(t time.Time) { b.now = t }

// --- Repeater scanner (chronic 0.10.2: regexp, first-match-wins) ---

var (
	seasonNameItems = []scanItem{
		item(regexp.MustCompile(`^springs?$`), "spring"),
		item(regexp.MustCompile(`^summers?$`), "summer"),
		item(regexp.MustCompile(`^(autumn)|(fall)s?$`), "autumn"),
		item(regexp.MustCompile(`^winters?$`), "winter"),
	}
	monthNameItems = []scanItem{
		item(regexp.MustCompile(`^jan[:.]?(uary)?$`), "january"),
		item(regexp.MustCompile(`^feb[:.]?(ruary)?$`), "february"),
		item(regexp.MustCompile(`^mar[:.]?(ch)?$`), "march"),
		item(regexp.MustCompile(`^apr[:.]?(il)?$`), "april"),
		item(regexp.MustCompile(`^may$`), "may"),
		item(regexp.MustCompile(`^jun[:.]?e?$`), "june"),
		item(regexp.MustCompile(`^jul[:.]?y?$`), "july"),
		item(regexp.MustCompile(`^aug[:.]?(ust)?$`), "august"),
		item(regexp.MustCompile(`^sep[:.]?(t[:.]?|tember)?$`), "september"),
		item(regexp.MustCompile(`^oct[:.]?(ober)?$`), "october"),
		item(regexp.MustCompile(`^nov[:.]?(ember)?$`), "november"),
		item(regexp.MustCompile(`^dec[:.]?(ember)?$`), "december"),
	}
	dayNameItems = []scanItem{
		item(regexp.MustCompile(`^m[ou]n(day)?$`), "monday"),
		item(regexp.MustCompile(`^t(ue|eu|oo|u|)s?(day)?$`), "tuesday"),
		item(regexp.MustCompile(`^we(d|dnes|nds|nns)(day)?$`), "wednesday"),
		item(regexp.MustCompile(`^th(u|ur|urs|ers)(day)?$`), "thursday"),
		item(regexp.MustCompile(`^fr[iy](day)?$`), "friday"),
		item(regexp.MustCompile(`^sat(t?[ue]rday)?$`), "saturday"),
		item(regexp.MustCompile(`^su[nm](day)?$`), "sunday"),
	}
	dayPortionItems = []scanItem{
		item(regexp.MustCompile(`^ams?$`), "am"),
		item(regexp.MustCompile(`^pms?$`), "pm"),
		item(regexp.MustCompile(`^mornings?$`), "morning"),
		item(regexp.MustCompile(`^afternoons?$`), "afternoon"),
		item(regexp.MustCompile(`^evenings?$`), "evening"),
		item(regexp.MustCompile(`^(night|nite)s?$`), "night"),
	}
	reTimeWord = regexp.MustCompile(`^\d{1,2}(:?\d{1,2})?([.:]?\d{1,2}([.:]\d{1,6})?)?$`)

	unitItems = []struct {
		re  *regexp.Regexp
		sym string
	}{
		{regexp.MustCompile(`^years?$`), "year"},
		{regexp.MustCompile(`^seasons?$`), "season"},
		{regexp.MustCompile(`^months?$`), "month"},
		{regexp.MustCompile(`^fortnights?$`), "fortnight"},
		{regexp.MustCompile(`^weeks?$`), "week"},
		{regexp.MustCompile(`^weekends?$`), "weekend"},
		{regexp.MustCompile(`^(week|business)days?$`), "weekday"},
		{regexp.MustCompile(`^days?$`), "day"},
		{regexp.MustCompile(`^hrs?$`), "hour"},
		{regexp.MustCompile(`^hours?$`), "hour"},
		{regexp.MustCompile(`^mins?$`), "minute"},
		{regexp.MustCompile(`^minutes?$`), "minute"},
		{regexp.MustCompile(`^secs?$`), "second"},
		{regexp.MustCompile(`^seconds?$`), "second"},
	}
)

// scanRepeater tags each token with at most one repeater, first matcher wins
// (season-name, month-name, day-name, day-portion, time, unit), mirroring
// Repeater.scan's `next` chain.
func scanRepeater(tokens []*token, opts *options) {
	for _, tok := range tokens {
		if t := scanSeasonName(tok); t != nil {
			tok.tag(t)
			continue
		}
		if t := scanMonthName(tok); t != nil {
			tok.tag(t)
			continue
		}
		if t := scanDayName(tok); t != nil {
			tok.tag(t)
			continue
		}
		if t := scanDayPortion(tok); t != nil {
			tok.tag(t)
			continue
		}
		if t := scanTimes(tok, opts); t != nil {
			tok.tag(t)
			continue
		}
		if t := scanUnits(tok, opts); t != nil {
			tok.tag(t)
			continue
		}
	}
}

func scanSeasonName(tok *token) *tag {
	if sym, ok := scanFor(tok.word, seasonNameItems); ok {
		return &tag{class: tcRepeaterSeasonName, sym: sym, rep: newRepeaterSeasonName(sym)}
	}
	return nil
}

func scanMonthName(tok *token) *tag {
	if sym, ok := scanFor(tok.word, monthNameItems); ok {
		return &tag{class: tcRepeaterMonthName, sym: sym, rep: newRepeaterMonthName(sym)}
	}
	return nil
}

func scanDayName(tok *token) *tag {
	if sym, ok := scanFor(tok.word, dayNameItems); ok {
		return &tag{class: tcRepeaterDayName, sym: sym, rep: newRepeaterDayName(sym)}
	}
	return nil
}

func scanDayPortion(tok *token) *tag {
	if sym, ok := scanFor(tok.word, dayPortionItems); ok {
		return &tag{class: tcRepeaterDayPortion, sym: sym, rep: newRepeaterDayPortion(sym, 0)}
	}
	return nil
}

func scanTimes(tok *token, opts *options) *tag {
	if reTimeWord.MatchString(tok.word) {
		return &tag{class: tcRepeaterTime, sym: tok.word, rep: newRepeaterTime(tok.word, opts)}
	}
	return nil
}

func scanUnits(tok *token, opts *options) *tag {
	for _, u := range unitItems {
		if u.re.MatchString(tok.word) {
			return newUnitRepeaterTag(u.sym, opts)
		}
	}
	return nil
}

// newUnitRepeaterTag builds the repeater tag for a time-unit word.
func newUnitRepeaterTag(sym string, opts *options) *tag {
	switch sym {
	case "year":
		return &tag{class: tcRepeaterYear, sym: sym, rep: &repeaterYear{}}
	case "season":
		return &tag{class: tcRepeaterSeason, sym: sym, rep: &repeaterSeason{}}
	case "month":
		return &tag{class: tcRepeaterMonth, sym: sym, rep: &repeaterMonth{}}
	case "fortnight":
		return &tag{class: tcRepeaterFortnight, sym: sym, rep: &repeaterFortnight{}}
	case "week":
		return &tag{class: tcRepeaterWeek, sym: sym, rep: &repeaterWeek{}}
	case "weekend":
		return &tag{class: tcRepeaterWeekend, sym: sym, rep: &repeaterWeekend{}}
	case "weekday":
		return &tag{class: tcRepeaterWeekday, sym: sym, rep: &repeaterWeekday{}}
	case "day":
		return &tag{class: tcRepeaterDay, sym: sym, rep: &repeaterDay{}}
	case "hour":
		return &tag{class: tcRepeaterHour, sym: sym, rep: &repeaterHour{}}
	case "minute":
		return &tag{class: tcRepeaterMinute, sym: sym, rep: &repeaterMinute{}}
	case "second":
		return &tag{class: tcRepeaterSecond, sym: sym, rep: &repeaterSecond{}}
	}
	return nil
}
