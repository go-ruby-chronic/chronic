// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"regexp"
	"strings"
)

// The pre-normalisation regexes for chronic 0.10.2, compiled once. They port
// Parser#pre_normalize verbatim.
var (
	reDotDate      = regexp.MustCompile(`\b(\d{2})\.(\d{2})\.(\d{4})\b`)
	reAmPmDot      = regexp.MustCompile(`\b([ap])\.m\.?`)
	reTzMinus      = regexp.MustCompile(`(\s+|:\d{2}|:\d{2}\.\d{3})\-(\d{2}:?\d{2})\b`)
	reAmPmColon    = regexp.MustCompile(`([ap]):m:?`)
	reQuotes       = regexp.MustCompile(`['"]`)
	reComma        = regexp.MustCompile(`,`)
	reSecondStart  = regexp.MustCompile(`^second `)
	reSecondOf     = regexp.MustCompile(`\bsecond (of|day|month|hour|minute|second)\b`)
	reSepPad       = regexp.MustCompile(`([/\-,@])`)
	reZeroPm       = regexp.MustCompile(`(?:^|\s)0(\d+:\d+\s*pm?\b)`)
	reToday        = regexp.MustCompile(`\btoday\b`)
	reTomorrow     = regexp.MustCompile(`\btomm?orr?ow\b`)
	reYesterday    = regexp.MustCompile(`\byesterday\b`)
	reNoon         = regexp.MustCompile(`\bnoon\b`)
	reMidnight     = regexp.MustCompile(`\bmidnight\b`)
	reNow          = regexp.MustCompile(`\bnow\b`)
	rePastTo       = regexp.MustCompile(`(\d{1,2}) (to|till|prior to|before)\b`)
	rePastAfter    = regexp.MustCompile(`(\d{1,2}) (after|past)\b`)
	reAgo          = regexp.MustCompile(`\b(?:ago|before(?: now)?)\b`)
	reThisLast     = regexp.MustCompile(`\bthis (?:last|past)\b`)
	reMorning      = regexp.MustCompile(`\b(?:in|during) the (morning)\b`)
	reDayparts     = regexp.MustCompile(`\b(?:in the|during the|at) (afternoon|evening|night)\b`)
	reTonight      = regexp.MustCompile(`\btonight\b`)
	reApShort      = regexp.MustCompile(`\b\d+:?\d*[ap]\b`)
	reFourDigitAp  = regexp.MustCompile(`\b(\d{2})(\d{2})(am|pm)\b`)
	reDigitApSpace = regexp.MustCompile(`(\d)([ap]m|oclock)\b`)
	reHence        = regexp.MustCompile(`\b(hence|after|from)\b`)
	reLeadingAn    = regexp.MustCompile(`(?i)^\s?an? `)
	reColonDate    = regexp.MustCompile(`\b(\d{4}):(\d{2}):(\d{2})\b`)
	reZeroHms      = regexp.MustCompile(`\b0(\d+):(\d{2}):(\d{2}) ([ap]m)\b`)
)

// preNormalize cleans up text ready for tokenising, porting chronic 0.10.2's
// Parser#pre_normalize. bias/nowYear are unused here (0.10.2 does not resolve
// quoted two-digit years during pre-normalisation) but kept for signature
// stability with the scanners.
func preNormalize(text string, bias, nowYear int) string {
	_ = bias
	_ = nowYear
	text = strings.ToLower(text)
	text = reDotDate.ReplaceAllString(text, `$3 / $2 / $1`)
	text = reAmPmDot.ReplaceAllString(text, `${1}m`)
	text = reTzMinus.ReplaceAllString(text, `${1}tzminus$2`)
	text = strings.ReplaceAll(text, ".", ":")
	text = reAmPmColon.ReplaceAllString(text, `${1}m`)
	text = reQuotes.ReplaceAllString(text, "")
	text = reComma.ReplaceAllString(text, " ")
	text = reSecondStart.ReplaceAllString(text, "2nd ")
	text = reSecondOf.ReplaceAllString(text, `2nd $1`)
	text = numerize(text)
	text = reSepPad.ReplaceAllString(text, ` $1 `)
	text = reZeroPm.ReplaceAllString(text, ` $1`)
	text = reToday.ReplaceAllString(text, "this day")
	text = reTomorrow.ReplaceAllString(text, "next day")
	text = reYesterday.ReplaceAllString(text, "last day")
	text = reNoon.ReplaceAllString(text, "12:00pm")
	text = reMidnight.ReplaceAllString(text, "24:00")
	text = reNow.ReplaceAllString(text, "this second")
	text = strings.ReplaceAll(text, "quarter", "15")
	text = strings.ReplaceAll(text, "half", "30")
	text = rePastTo.ReplaceAllString(text, `$1 minutes past`)
	text = rePastAfter.ReplaceAllString(text, `$1 minutes future`)
	text = reAgo.ReplaceAllString(text, "past")
	text = reThisLast.ReplaceAllString(text, "last")
	text = reMorning.ReplaceAllString(text, `$1`)
	text = reDayparts.ReplaceAllString(text, `$1`)
	text = reTonight.ReplaceAllString(text, "this night")
	text = reApShort.ReplaceAllString(text, `${0}m`)
	text = reFourDigitAp.ReplaceAllString(text, `$1:$2$3`)
	text = reDigitApSpace.ReplaceAllString(text, `$1 $2`)
	text = reHence.ReplaceAllString(text, "future")
	text = reLeadingAn.ReplaceAllString(text, "1 ")
	text = reColonDate.ReplaceAllString(text, `$1 / $2 / $3`)
	text = reZeroHms.ReplaceAllString(text, `$1:$2:$3 $4`)
	return text
}
