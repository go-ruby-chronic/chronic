// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

// tagClass enumerates every concrete Chronic tag class. The handler matcher and
// handlers look tags up by class, so each Ruby tag class gets one constant. The
// grammar also matches against a few *abstract* base classes (repeater, scalar,
// ordinal, separator, sign, pointer, grabber, timezone); isA encodes that
// subclass relationship.
type tagClass int

const (
	tcNone tagClass = iota

	tcGrabber
	tcPointer
	tcSign
	tcSignPlus
	tcSignMinus
	tcTimeZone

	tcOrdinal
	tcOrdinalDay
	tcOrdinalMonth
	tcOrdinalYear

	tcScalar
	tcScalarWide
	tcScalarSubsecond
	tcScalarSecond
	tcScalarMinute
	tcScalarHour
	tcScalarDay
	tcScalarMonth
	tcScalarYear

	tcSeparator
	tcSeparatorComma
	tcSeparatorDot
	tcSeparatorColon
	tcSeparatorSpace
	tcSeparatorSlash
	tcSeparatorDash
	tcSeparatorQuote
	tcSeparatorAt
	tcSeparatorIn
	tcSeparatorOn
	tcSeparatorAnd
	tcSeparatorT
	tcSeparatorW

	tcRepeater
	tcRepeaterYear
	tcRepeaterSeason
	tcRepeaterSeasonName
	tcRepeaterMonth
	tcRepeaterMonthName
	tcRepeaterFortnight
	tcRepeaterWeek
	tcRepeaterWeekend
	tcRepeaterWeekday
	tcRepeaterDay
	tcRepeaterDayName
	tcRepeaterDayPortion
	tcRepeaterHour
	tcRepeaterMinute
	tcRepeaterSecond
	tcRepeaterTime
)

// baseOf maps a concrete class to its abstract base for is-a matching, or tcNone
// when the class is itself only matched concretely.
var baseOf = map[tagClass]tagClass{
	tcSignPlus:  tcSign,
	tcSignMinus: tcSign,

	tcOrdinalDay:   tcOrdinal,
	tcOrdinalMonth: tcOrdinal,
	tcOrdinalYear:  tcOrdinal,

	tcScalarWide:      tcScalar,
	tcScalarSubsecond: tcScalar,
	tcScalarSecond:    tcScalar,
	tcScalarMinute:    tcScalar,
	tcScalarHour:      tcScalar,
	tcScalarDay:       tcScalar,
	tcScalarMonth:     tcScalar,
	tcScalarYear:      tcScalar,

	tcSeparatorComma: tcSeparator,
	tcSeparatorDot:   tcSeparator,
	tcSeparatorColon: tcSeparator,
	tcSeparatorSpace: tcSeparator,
	tcSeparatorSlash: tcSeparator,
	tcSeparatorDash:  tcSeparator,
	tcSeparatorQuote: tcSeparator,
	tcSeparatorAt:    tcSeparator,
	tcSeparatorIn:    tcSeparator,
	tcSeparatorOn:    tcSeparator,
	tcSeparatorAnd:   tcSeparator,
	tcSeparatorT:     tcSeparator,
	tcSeparatorW:     tcSeparator,

	tcRepeaterYear:        tcRepeater,
	tcRepeaterSeason:      tcRepeater,
	tcRepeaterSeasonName:  tcRepeater,
	tcRepeaterMonth:       tcRepeater,
	tcRepeaterMonthName:   tcRepeater,
	tcRepeaterFortnight:   tcRepeater,
	tcRepeaterWeek:        tcRepeater,
	tcRepeaterWeekend:     tcRepeater,
	tcRepeaterWeekday:     tcRepeater,
	tcRepeaterDay:         tcRepeater,
	tcRepeaterDayName:     tcRepeater,
	tcRepeaterDayPortion:  tcRepeater,
	tcRepeaterHour:        tcRepeater,
	tcRepeaterMinute:      tcRepeater,
	tcRepeaterSecond:      tcRepeater,
	tcRepeaterTime:        tcRepeater,
	// Season/Quarter *name* repeaters also inherit their non-name repeater.
}

// nameParent maps *name repeaters to the plain repeater whose implementation
// they reuse (RepeaterSeasonName < RepeaterSeason, RepeaterQuarterName <
// RepeaterQuarter). Matching only needs the tcRepeater base above, but the
// implementations share behaviour via the season/quarter engines directly.

// tag is a token annotation. Concrete tags carry a class, an optional symbolic
// type (kind), an integer value (for scalars/ordinals), and a repeater engine
// when the tag participates in the repeater pipeline.
type tag struct {
	class tagClass
	sym   string // symbolic type, e.g. "monday", "am", "this", "future"
	num   int    // numeric type for scalars/ordinals/quarter values
	width int
	rep   repeater // non-nil for repeater tags
}

// isA reports whether the tag is an instance of (or subclass of) class c.
func (t *tag) isA(c tagClass) bool {
	for k := t.class; k != tcNone; k = baseOf[k] {
		if k == c {
			return true
		}
	}
	return false
}

// token is a tokenised word plus the tags scanners attached to it.
type token struct {
	word string
	tags []*tag
}

func newToken(word string) *token {
	return &token{word: word}
}

// tag attaches a tag (ignoring nil), mirroring Token#tag.
func (t *token) tag(nt *tag) {
	if nt != nil {
		t.tags = append(t.tags, nt)
	}
}

// untag removes every tag of the given class, mirroring Token#untag.
func (t *token) untag(c tagClass) {
	out := t.tags[:0]
	for _, m := range t.tags {
		if !m.isA(c) {
			out = append(out, m)
		}
	}
	t.tags = out
}

// tagged reports whether the token carries any tag.
func (t *token) tagged() bool { return len(t.tags) > 0 }

// getTag returns the first tag that is-a c, or nil.
func (t *token) getTag(c tagClass) *tag {
	for _, m := range t.tags {
		if m.isA(c) {
			return m
		}
	}
	return nil
}
