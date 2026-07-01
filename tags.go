// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "regexp"

// This file ports chronic 0.10.2's simple (non-repeater) tag scanners: Grabber,
// Pointer, Sign, TimeZone, Separator, Ordinal and Scalar. Every scanner matches
// by regexp against token.word and stops at the first match within the scanner.

// --- Grabber ---

var grabberItems = []scanItem{
	item(regexp.MustCompile(`last`), "last"),
	item(regexp.MustCompile(`this`), "this"),
	item(regexp.MustCompile(`next`), "next"),
}

func scanGrabber(tokens []*token) {
	for _, tok := range tokens {
		if sym, ok := scanFor(tok.word, grabberItems); ok {
			tok.tag(&tag{class: tcGrabber, sym: sym})
		}
	}
}

// --- Pointer ---

var pointerItems = []scanItem{
	item(regexp.MustCompile(`\bpast\b`), "past"),
	item(regexp.MustCompile(`\b(?:future|in)\b`), "future"),
}

func scanPointer(tokens []*token) {
	for _, tok := range tokens {
		if sym, ok := scanFor(tok.word, pointerItems); ok {
			tok.tag(&tag{class: tcPointer, sym: sym})
		}
	}
}

// --- Sign ---

var (
	reSignPlus  = regexp.MustCompile(`^\+$`)
	reSignMinus = regexp.MustCompile(`^-$`)
)

func scanSign(tokens []*token) {
	for _, tok := range tokens {
		if reSignPlus.MatchString(tok.word) {
			tok.tag(&tag{class: tcSignPlus, sym: "plus"})
			continue
		}
		if reSignMinus.MatchString(tok.word) {
			tok.tag(&tag{class: tcSignMinus, sym: "minus"})
		}
	}
}

// --- TimeZone ---

var (
	reTzWord   = regexp.MustCompile(`(?i)[PMCE][DS]T|UTC`)
	reTzOffset = regexp.MustCompile(`(tzminus)?\d{2}:?\d{2}`)
)

func scanTimeZone(tokens []*token) {
	for _, tok := range tokens {
		if reTzWord.MatchString(tok.word) || reTzOffset.MatchString(tok.word) {
			tok.tag(&tag{class: tcTimeZone, sym: "tz"})
		}
	}
}

// --- Separator ---

var separatorScanners = []struct {
	re    *regexp.Regexp
	class tagClass
	sym   string
}{
	{regexp.MustCompile(`^,$`), tcSeparatorComma, "comma"},
	{regexp.MustCompile(`^\.$`), tcSeparatorDot, "dot"},
	{regexp.MustCompile(`^:$`), tcSeparatorColon, "colon"},
	{regexp.MustCompile(`^ $`), tcSeparatorSpace, "space"},
	{regexp.MustCompile(`^/$`), tcSeparatorSlash, "slash"},
	{regexp.MustCompile(`^-$`), tcSeparatorDash, "dash"},
	{regexp.MustCompile(`^'$`), tcSeparatorQuote, "single_quote"},
	{regexp.MustCompile(`^"$`), tcSeparatorQuote, "double_quote"},
	{regexp.MustCompile(`^(at|@)$`), tcSeparatorAt, "at"},
	{regexp.MustCompile(`^in$`), tcSeparatorIn, "in"},
	{regexp.MustCompile(`^on$`), tcSeparatorOn, "on"},
	{regexp.MustCompile(`^and$`), tcSeparatorAnd, "and"},
	{regexp.MustCompile(`^t$`), tcSeparatorT, "T"},
	{regexp.MustCompile(`^w$`), tcSeparatorW, "W"},
}

func scanSeparator(tokens []*token) {
	for _, tok := range tokens {
		for _, s := range separatorScanners {
			if s.re.MatchString(tok.word) {
				tok.tag(&tag{class: s.class, sym: s.sym})
				break
			}
		}
	}
}

// --- Ordinal ---

var reOrdinal = regexp.MustCompile(`^(\d+)(st|nd|rd|th|\.)$`)

func scanOrdinal(tokens []*token, opts *options) {
	for _, tok := range tokens {
		if m := reOrdinal.FindStringSubmatch(tok.word); m != nil {
			ordinal := atoi(m[1])
			tok.tag(&tag{class: tcOrdinal, num: ordinal})
			if couldBeDay(ordinal) {
				tok.tag(&tag{class: tcOrdinalDay, num: ordinal})
			}
			if couldBeMonth(ordinal) {
				tok.tag(&tag{class: tcOrdinalMonth, num: ordinal})
			}
			if couldBeYear(ordinal) {
				year := makeYear(ordinal, opts.ambiguousYearFutureBias, opts.now.Year())
				tok.tag(&tag{class: tcOrdinalYear, num: year})
			}
		}
	}
}

// --- Scalar ---

var (
	reDigits      = regexp.MustCompile(`^\d+$`)
	dayPortionSet = map[string]bool{
		"am": true, "pm": true, "morning": true,
		"afternoon": true, "evening": true, "night": true,
	}
)

func scanScalar(tokens []*token, opts *options) {
	for i, tok := range tokens {
		var postWord string
		if i+1 < len(tokens) {
			postWord = tokens[i+1].word
		}
		if !reDigits.MatchString(tok.word) {
			continue
		}
		scalar := atoi(tok.word)
		tok.tag(&tag{class: tcScalar, num: scalar})
		if couldBeSubsecond(scalar) {
			tok.tag(&tag{class: tcScalarSubsecond, num: scalar})
		}
		if couldBeSecond(scalar) {
			tok.tag(&tag{class: tcScalarSecond, num: scalar})
		}
		if couldBeMinute(scalar) {
			tok.tag(&tag{class: tcScalarMinute, num: scalar})
		}
		if couldBeHour(scalar, opts.hours24Set && !opts.hours24) {
			tok.tag(&tag{class: tcScalarHour, num: scalar})
		}
		if !(postWord != "" && dayPortionSet[postWord]) {
			if couldBeDay(scalar) {
				tok.tag(&tag{class: tcScalarDay, num: scalar})
			}
			if couldBeMonth(scalar) {
				tok.tag(&tag{class: tcScalarMonth, num: scalar})
			}
			if couldBeYear(scalar) {
				year := makeYear(scalar, opts.ambiguousYearFutureBias, opts.now.Year())
				tok.tag(&tag{class: tcScalarYear, num: year})
			}
		}
	}
}
