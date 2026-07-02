// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"sort"
	"time"
)

// This file ports Chronic::Handlers — the methods that turn matched token
// sequences into Spans. Each handler mirrors its Ruby namesake.

// --- month/day core ---

// handleMD ports handle_m_d. month is any repeater whose this() yields the
// target month span (RepeaterMonthName for named months, RepeaterMonth for
// "this/next month"). 0.10.2's handle_m_d does no future/past rollover.
func (p *parser) handleMD(month repeater, day int, timeTokens []*token, opts *options) *Span {
	month.setStart(p.now)
	span := month.this(opts.context)
	year, mon := span.Begin.Year(), int(span.Begin.Month())
	dayStart := local(year, mon, day)
	return p.dayOrTime(dayStart, timeTokens, opts)
}

func handleRmnSd(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[1].getTag(tcScalarDay).num
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tokens[2:], opts)
}

func handleRmnSdOn(p *parser, tokens []*token, opts *options) *Span {
	var month *repeaterMonthName
	var day int
	var tr []*token
	if len(tokens) > 3 {
		month = tokens[2].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
		day = tokens[3].getTag(tcScalarDay).num
		tr = tokens[0:2]
	} else {
		month = tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
		day = tokens[2].getTag(tcScalarDay).num
		tr = tokens[0:1]
	}
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tr, opts)
}

func handleRmnOd(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[1].getTag(tcOrdinalDay).num
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tokens[2:], opts)
}

func handleOdRm(p *parser, tokens []*token, opts *options) *Span {
	day := tokens[0].getTag(tcOrdinalDay).num
	month := tokens[2].getTag(tcRepeaterMonth).rep
	return p.handleMD(month, day, tokens[3:], opts)
}

func handleOdRmn(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[0].getTag(tcOrdinalDay).num
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tokens[2:], opts)
}

func handleSyRmnOd(p *parser, tokens []*token, opts *options) *Span {
	year := tokens[0].getTag(tcScalarYear).num
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	day := tokens[2].getTag(tcOrdinalDay).num
	timeTokens := lastN(tokens, len(tokens)-3)
	if monthOverflow(year, month, day) {
		return nil
	}
	dayStart := local(year, month, day)
	return p.dayOrTime(dayStart, timeTokens, opts)
}

func handleSdRmn(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[0].getTag(tcScalarDay).num
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tokens[2:], opts)
}

func handleRmnOdOn(p *parser, tokens []*token, opts *options) *Span {
	var month *repeaterMonthName
	var day int
	var tr []*token
	if len(tokens) > 3 {
		month = tokens[2].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
		day = tokens[3].getTag(tcOrdinalDay).num
		tr = tokens[0:2]
	} else {
		month = tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
		day = tokens[2].getTag(tcOrdinalDay).num
		tr = tokens[0:1]
	}
	if monthOverflow(p.now.Year(), month.index(), day) {
		return nil
	}
	return p.handleMD(month, day, tr, opts)
}

func handleRmnSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	year := tokens[1].getTag(tcScalarYear).num
	return p.handleYearAndMonth(year, month)
}

func handleGeneric(p *parser, tokens []*token, opts *options) *Span {
	t, ok := parseGenericTimestamp(opts.text)
	if !ok {
		return nil
	}
	return newSpan(t, t.Add(time.Second))
}

func handleRmnSdSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	day := tokens[1].getTag(tcScalarDay).num
	year := tokens[2].getTag(tcScalarYear).num
	timeTokens := lastN(tokens, len(tokens)-3)
	if monthOverflow(year, month, day) {
		return nil
	}
	return p.dayOrTime(local(year, month, day), timeTokens, opts)
}

func handleRmnOdSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	day := tokens[1].getTag(tcOrdinalDay).num
	year := tokens[2].getTag(tcScalarYear).num
	timeTokens := lastN(tokens, len(tokens)-3)
	if monthOverflow(year, month, day) {
		return nil
	}
	return p.dayOrTime(local(year, month, day), timeTokens, opts)
}

func handleOdRmnSy(p *parser, tokens []*token, opts *options) *Span {
	day := tokens[0].getTag(tcOrdinalDay).num
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	year := tokens[2].getTag(tcScalarYear).num
	timeTokens := lastN(tokens, len(tokens)-3)
	if monthOverflow(year, month, day) {
		return nil
	}
	return p.dayOrTime(local(year, month, day), timeTokens, opts)
}

func handleSdRmnSy(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[0], tokens[2]}
	timeTokens := lastN(tokens, len(tokens)-3)
	return handleRmnSdSy(p, append(newTokens, timeTokens...), opts)
}

func handleSmSdSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcScalarMonth).num
	day := tokens[1].getTag(tcScalarDay).num
	year := tokens[2].getTag(tcScalarYear).num
	timeTokens := lastN(tokens, len(tokens)-3)
	if monthOverflow(year, month, day) {
		return nil
	}
	return p.dayOrTime(local(year, month, day), timeTokens, opts)
}

func handleSdSmSy(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[0], tokens[2]}
	timeTokens := lastN(tokens, len(tokens)-3)
	return handleSmSdSy(p, append(newTokens, timeTokens...), opts)
}

func handleSySmSd(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[2], tokens[0]}
	timeTokens := lastN(tokens, len(tokens)-3)
	return handleSmSdSy(p, append(newTokens, timeTokens...), opts)
}

func handleSmSd(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcScalarMonth).num
	day := tokens[1].getTag(tcScalarDay).num
	year := p.now.Year()
	timeTokens := lastN(tokens, len(tokens)-2)
	if monthOverflow(year, month, day) {
		return nil
	}
	dayStart := local(year, month, day)
	if opts.context == "future" && dayStart.Before(p.now) {
		dayStart = local(year+1, month, day)
	}
	return p.dayOrTime(dayStart, timeTokens, opts)
}

func handleSdSm(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[0]}
	timeTokens := lastN(tokens, len(tokens)-2)
	return handleSmSd(p, append(newTokens, timeTokens...), opts)
}

func (p *parser) handleYearAndMonth(year, month int) *Span {
	var ny, nm int
	if month == 12 {
		ny, nm = year+1, 1
	} else {
		ny, nm = year, month+1
	}
	return newSpan(local(year, month), local(ny, nm))
}

func handleSmSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[0].getTag(tcScalarMonth).num
	year := tokens[1].getTag(tcScalarYear).num
	return p.handleYearAndMonth(year, month)
}

func handleSySm(p *parser, tokens []*token, opts *options) *Span {
	year := tokens[0].getTag(tcScalarYear).num
	month := tokens[1].getTag(tcScalarMonth).num
	return p.handleYearAndMonth(year, month)
}

// --- day-name led forms ---

func handleRdnRmnOd(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[2].getTag(tcOrdinalDay).num
	timeTokens := lastN(tokens, len(tokens)-3)
	year := p.now.Year()
	if monthOverflow(year, month.index(), day) {
		return nil
	}
	return p.dnBody(year, month.index(), day, timeTokens, opts)
}

func handleRdnRmnOdSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[2].getTag(tcOrdinalDay).num
	year := tokens[3].getTag(tcScalarYear).num
	if monthOverflow(year, month.index(), day) {
		return nil
	}
	return newSpan(local(year, month.index(), day), p.timeWithRollover(year, month.index(), day+1))
}

func handleRdnOd(p *parser, tokens []*token, opts *options) *Span {
	day := tokens[1].getTag(tcOrdinalDay).num
	timeTokens := lastN(tokens, len(tokens)-2)
	year := p.now.Year()
	month := int(p.now.Month())
	if opts.context == "future" {
		if p.now.Day() > day {
			month++
		}
	}
	if monthOverflow(year, month, day) {
		return nil
	}
	return p.dnBody(year, month, day, timeTokens, opts)
}

func handleRdnRmnSd(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[2].getTag(tcScalarDay).num
	timeTokens := lastN(tokens, len(tokens)-3)
	year := p.now.Year()
	if monthOverflow(year, month.index(), day) {
		return nil
	}
	return p.dnBody(year, month.index(), day, timeTokens, opts)
}

func handleRdnRmnSdSy(p *parser, tokens []*token, opts *options) *Span {
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName)
	day := tokens[2].getTag(tcScalarDay).num
	year := tokens[3].getTag(tcScalarYear).num
	if monthOverflow(year, month.index(), day) {
		return nil
	}
	return newSpan(local(year, month.index(), day), p.timeWithRollover(year, month.index(), day+1))
}

// dnBody is the common tail of the day-name handlers: either a whole-day span or
// a time within the day.
func (p *parser) dnBody(year, month, day int, timeTokens []*token, opts *options) *Span {
	if len(timeTokens) == 0 {
		return newSpan(local(year, month, day), p.timeWithRollover(year, month, day+1))
	}
	return p.dayOrTime(local(year, month, day), timeTokens, opts)
}

func handleSmRmnSy(p *parser, tokens []*token, opts *options) *Span {
	day := tokens[0].getTag(tcScalarDay).num
	month := tokens[1].getTag(tcRepeaterMonthName).rep.(*repeaterMonthName).index()
	year := tokens[2].getTag(tcScalarYear).num
	if len(tokens) > 3 {
		anchorSpan := p.getAnchor([]*token{tokens[len(tokens)-1]}, opts)
		t := anchorSpan.Begin
		h, m, s := t.Hour(), t.Minute(), t.Second()
		start := local(year, month, day, h, m, s)
		end := local(year, month, day+1, h, m, s)
		return newSpan(start, end)
	}
	t := local(year, month, day)
	if day < 31 {
		day++
	}
	return newSpan(t, local(year, month, day))
}

// --- anchors ---

func handleR(p *parser, tokens []*token, opts *options) *Span {
	dd := p.dealiasAndDisambiguateTimes(tokens, opts)
	return p.getAnchor(dd, opts)
}

func handleRGR(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[0], tokens[2]}
	return handleR(p, newTokens, opts)
}

// --- arrows ---

func (p *parser) handleSRP(tokens []*token, span *Span, opts *options) *Span {
	distance := tokens[0].getTag(tcScalar).num
	rep := tokens[1].getTag(tcRepeater).rep
	pointer := tokens[2].getTag(tcPointer).sym
	if o, ok := rep.(offsetter); ok {
		return o.offset(span, distance, pointer)
	}
	return nil
}

func handleSRP(p *parser, tokens []*token, opts *options) *Span {
	span := newSpan(p.now, p.now.Add(time.Second))
	return p.handleSRP(tokens, span, opts)
}

func handlePSR(p *parser, tokens []*token, opts *options) *Span {
	newTokens := []*token{tokens[1], tokens[2], tokens[0]}
	return handleSRP(p, newTokens, opts)
}

func handleSRPA(p *parser, tokens []*token, opts *options) *Span {
	anchorSpan := p.getAnchor(tokens[3:], opts)
	if anchorSpan == nil {
		return nil
	}
	return p.handleSRP(tokens, anchorSpan, opts)
}

func handleSRASRPA(p *parser, tokens []*token, opts *options) *Span {
	anchorSpan := p.getAnchor(tokens[4:], opts)
	if anchorSpan == nil {
		return nil
	}
	span := p.handleSRP(append(append([]*token{}, tokens[0:2]...), tokens[4:7]...), anchorSpan, opts)
	if span == nil {
		return nil
	}
	return p.handleSRP(append(append([]*token{}, tokens[2:4]...), tokens[4:7]...), span, opts)
}

// --- narrows ---

func (p *parser) handleORR(tokens []*token, outerSpan *Span, opts *options) *Span {
	rep := tokens[1].getTag(tcRepeater).rep
	rep.setStart(outerSpan.Begin.Add(-time.Second))
	ordinal := tokens[0].getTag(tcOrdinal).num
	var span *Span
	for i := 0; i < ordinal; i++ {
		span = rep.next("future")
		if !span.Begin.Before(outerSpan.End) {
			span = nil
			break
		}
	}
	return span
}

func handleORSR(p *parser, tokens []*token, opts *options) *Span {
	// The anchor is a single repeater token, which getAnchor always resolves to a
	// non-nil span (the gem's `return unless outer_span` guarded a dynamically-nil
	// case unreachable through this fixed single-token pattern).
	outerSpan := p.getAnchor([]*token{tokens[3]}, opts)
	return p.handleORR(tokens[0:2], outerSpan, opts)
}

func handleORGR(p *parser, tokens []*token, opts *options) *Span {
	outerSpan := p.getAnchor(tokens[2:4], opts)
	if outerSpan == nil {
		return nil
	}
	return p.handleORR(tokens[0:2], outerSpan, opts)
}

// --- support ---

func (p *parser) dayOrTime(dayStart time.Time, timeTokens []*token, opts *options) *Span {
	outerSpan := newSpan(dayStart, dayStart.Add(daySeconds*time.Second))
	if len(timeTokens) == 0 {
		return outerSpan
	}
	p.now = outerSpan.Begin
	futureOpts := *opts
	futureOpts.context = "future"
	return p.getAnchor(p.dealiasAndDisambiguateTimes(timeTokens, opts), &futureOpts)
}

func (p *parser) getAnchor(tokens []*token, opts *options) *Span {
	grabber := "this"
	pointer := "future"
	repeaters := getRepeaters(tokens)
	tokens = tokens[:len(tokens)-len(repeaters)]

	if len(tokens) > 0 && tokens[0].getTag(tcGrabber) != nil {
		grabber = tokens[0].getTag(tcGrabber).sym
		tokens = tokens[1:]
	}
	_ = tokens

	head := repeaters[0]
	repeaters = repeaters[1:]
	head.setStart(p.now)

	var outerSpan *Span
	switch grabber {
	case "last":
		outerSpan = head.next("past")
	case "this":
		if opts.context != "past" && len(repeaters) > 0 {
			outerSpan = head.this("none")
		} else {
			outerSpan = head.this(opts.context)
		}
	case "next":
		outerSpan = head.next("future")
	}
	if outerSpan == nil {
		return nil
	}
	return findWithin(repeaters, outerSpan, pointer)
}

// getRepeaters extracts repeater engines from tokens, sorted by width
// descending (widest first), mirroring get_repeaters' sort.reverse.
func getRepeaters(tokens []*token) []repeater {
	var reps []repeater
	for _, t := range tokens {
		if rt := t.getTag(tcRepeater); rt != nil {
			reps = append(reps, rt.rep)
		}
	}
	sort.SliceStable(reps, func(i, j int) bool {
		return reps[i].width() > reps[j].width()
	})
	return reps
}

func monthOverflow(year, month, day int) bool {
	// The bounds guard makes the monthDays index always valid, so no recover is
	// needed here (the gem's rescue guarded a different, dynamically-typed path).
	if month < 1 || month > 12 {
		return false
	}
	if isLeap(year) {
		return day > monthDaysLeap[month]
	}
	return day > monthDays[month]
}

// findWithin recursively narrows repeaters within a span, mirroring find_within.
func findWithin(tags []repeater, span *Span, pointer string) *Span {
	if len(tags) == 0 {
		return span
	}
	head := tags[0]
	rest := tags[1:]
	if pointer == "future" {
		head.setStart(span.Begin)
	} else {
		head.setStart(span.End)
	}
	h := head.this("none")
	if span.cover(h.Begin) || span.cover(h.End) {
		return findWithin(rest, h, pointer)
	}
	return nil
}

func (p *parser) timeWithRollover(year, month, day int) time.Time {
	if monthOverflow(year, month, day) {
		if month == 12 {
			return local(year+1, 1, 1)
		}
		return local(year, month+1, 1)
	}
	return local(year, month, day)
}

// dealiasAndDisambiguateTimes ports the same-named support method.
func (p *parser) dealiasAndDisambiguateTimes(tokens []*token, opts *options) []*token {
	dayPortionIndex := -1
	for i, t := range tokens {
		if t.getTag(tcRepeaterDayPortion) != nil {
			dayPortionIndex = i
			break
		}
	}
	timeIndex := -1
	for i, t := range tokens {
		if t.getTag(tcRepeaterTime) != nil {
			timeIndex = i
			break
		}
	}

	if dayPortionIndex >= 0 && timeIndex >= 0 {
		t1 := tokens[dayPortionIndex]
		t1tag := t1.getTag(tcRepeaterDayPortion)
		switch t1tag.sym {
		case "morning":
			t1.untag(tcRepeaterDayPortion)
			t1.tag(&tag{class: tcRepeaterDayPortion, sym: "am", rep: newRepeaterDayPortion("am", 0)})
		case "afternoon", "evening", "night":
			t1.untag(tcRepeaterDayPortion)
			t1.tag(&tag{class: tcRepeaterDayPortion, sym: "pm", rep: newRepeaterDayPortion("pm", 0)})
		}
	}

	if opts.ambiguousTimeRange != -1 {
		var out []*token
		for i, tok := range tokens {
			out = append(out, tok)
			var nextTok *token
			if i+1 < len(tokens) {
				nextTok = tokens[i+1]
			}
			rt := tok.getTag(tcRepeaterTime)
			if rt != nil && rt.rep.(*repeaterTime).ambiguous() &&
				(nextTok == nil || nextTok.getTag(tcRepeaterDayPortion) == nil) {
				dis := newToken("disambiguator")
				dis.tag(&tag{class: tcRepeaterDayPortion, sym: "", rep: newRepeaterDayPortion("", opts.ambiguousTimeRange)})
				out = append(out, dis)
			}
		}
		tokens = out
	}
	return tokens
}

// lastN returns the last n tokens (or all when n>=len, or none when n<=0),
// mirroring Array#last(n).
func lastN(tokens []*token, n int) []*token {
	if n <= 0 {
		return nil
	}
	if n >= len(tokens) {
		return tokens
	}
	return tokens[len(tokens)-n:]
}
