// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"regexp"
	"strconv"
	"strings"
)

// numerize converts English number words to digits, porting chronic 0.10.2's
// bundled Chronic::Numerizer (NOT the standalone `numerizer` gem). Chronic calls
// it during pre-normalisation so "three days ago" -> "3 days ago" and "one
// hundred and thirty six days from now" -> "136 days ...". Notably the 0.10.2
// ORDINALS table omits "second" so the time unit survives.
func numerize(s string) string {
	// preprocess: collapse spaces; split hyphenated non-digit words.
	s = rePreSpace.ReplaceAllStringFunc(s, func(m string) string {
		sub := rePreSpace.FindStringSubmatch(m)
		if sub[1] != "" || sub[2] != "" {
			return sub[1] + " " + sub[2]
		}
		return " "
	})
	// take the 'a' out of "a half" so it doesn't become 1; restore at the end.
	s = regexp.MustCompile(`a half`).ReplaceAllString(s, "haAlf")

	for _, dn := range directNums {
		s = dn.re.ReplaceAllString(s, "<num>"+dn.rep)
	}
	for _, on := range ordinalNums {
		s = on.re.ReplaceAllString(s, "<num>"+on.rep)
	}
	// ten-prefix + following single digit => sum
	for _, tp := range tenPrefixes {
		re := regexp.MustCompile(`(?i)(?:` + tp.word + `) *<num>(\d(?:[^\d]|$))?`)
		s = gsubFunc(re, s, func(g []string) string {
			d := 0
			if g[1] != "" {
				// g[1] may carry the trailing boundary char; take leading digit.
				d = atoi(g[1][:1])
			}
			return "<num>" + strconv.Itoa(tp.val+d)
		})
	}
	for _, tp := range tenPrefixes {
		re := regexp.MustCompile(`(?i)` + tp.word)
		s = re.ReplaceAllString(s, "<num>"+strconv.Itoa(tp.val))
	}
	for _, bp := range bigPrefixes {
		re := regexp.MustCompile(`(?i)(?:<num>)?(\d*) *` + bp.word)
		s = gsubFunc(re, s, func(g []string) string {
			if g[1] == "" {
				return strconv.Itoa(bp.val)
			}
			return "<num>" + strconv.Itoa(bp.val*atoi(g[1]))
		})
		s = andition(s)
	}
	// fractional addition for the stashed "a half".
	reHalf := regexp.MustCompile(`(?i)(\d+)(?: | and |-)*haAlf`)
	s = gsubFunc(reHalf, s, func(g []string) string {
		return strconv.FormatFloat(float64(atoi(g[1]))+0.5, 'f', -1, 64)
	})

	return strings.ReplaceAll(s, "<num>", "")
}

var rePreSpace = regexp.MustCompile(` +|([^\d])-([^\d])`)

// directNums / ordinalNums pair a compiled matcher with its replacement, in the
// gem's order. The "four(\W|$)"-style entries guard four/six/… so they do not
// match fourty/sixty; the replacement re-emits the boundary via $1.
var directNums = []struct {
	re  *regexp.Regexp
	rep string
}{
	{regexp.MustCompile(`(?i)eleven`), "11"},
	{regexp.MustCompile(`(?i)twelve`), "12"},
	{regexp.MustCompile(`(?i)thirteen`), "13"},
	{regexp.MustCompile(`(?i)fourteen`), "14"},
	{regexp.MustCompile(`(?i)fifteen`), "15"},
	{regexp.MustCompile(`(?i)sixteen`), "16"},
	{regexp.MustCompile(`(?i)seventeen`), "17"},
	{regexp.MustCompile(`(?i)eighteen`), "18"},
	{regexp.MustCompile(`(?i)nineteen`), "19"},
	{regexp.MustCompile(`(?i)ninteen`), "19"},
	{regexp.MustCompile(`(?i)zero`), "0"},
	{regexp.MustCompile(`(?i)one`), "1"},
	{regexp.MustCompile(`(?i)two`), "2"},
	{regexp.MustCompile(`(?i)four(\W|$)`), "4$1"},
	{regexp.MustCompile(`(?i)five`), "5"},
	{regexp.MustCompile(`(?i)six(\W|$)`), "6$1"},
	{regexp.MustCompile(`(?i)seven(\W|$)`), "7$1"},
	{regexp.MustCompile(`(?i)eight(\W|$)`), "8$1"},
	{regexp.MustCompile(`(?i)nine(\W|$)`), "9$1"},
	{regexp.MustCompile(`(?i)ten`), "10"},
	// Ruby's /\ba[\b^$]/ has a char class of {backspace, ^, $} — literal control
	// characters that never appear in Chronic input, so this entry is effectively
	// inert. We reproduce it with the backspace/caret/dollar class (RE2-safe).
	{regexp.MustCompile("(?i)\\ba[\\x08^$]"), "1"},
	{regexp.MustCompile(`(?i)three`), "3"},
}

// ordinalNums appends the last two characters of the word (Ruby's on[0][-2,2])
// to keep the ordinal suffix, e.g. "third" -> "<num>3rd". "second" is
// intentionally absent so the time unit is preserved.
var ordinalNums = []struct {
	re  *regexp.Regexp
	rep string
}{
	{regexp.MustCompile(`(?i)first`), "1st"},
	{regexp.MustCompile(`(?i)third`), "3rd"},
	{regexp.MustCompile(`(?i)fourth`), "4th"},
	{regexp.MustCompile(`(?i)fifth`), "5th"},
	{regexp.MustCompile(`(?i)sixth`), "6th"},
	{regexp.MustCompile(`(?i)seventh`), "7th"},
	{regexp.MustCompile(`(?i)eighth`), "8th"},
	{regexp.MustCompile(`(?i)ninth`), "9th"},
	{regexp.MustCompile(`(?i)tenth`), "10th"},
	{regexp.MustCompile(`(?i)twelfth`), "12th"},
	{regexp.MustCompile(`(?i)twentieth`), "20th"},
	{regexp.MustCompile(`(?i)thirtieth`), "30th"},
	{regexp.MustCompile(`(?i)fourtieth`), "40th"},
	{regexp.MustCompile(`(?i)fiftieth`), "50th"},
	{regexp.MustCompile(`(?i)sixtieth`), "60th"},
	{regexp.MustCompile(`(?i)seventieth`), "70th"},
	{regexp.MustCompile(`(?i)eightieth`), "80th"},
	{regexp.MustCompile(`(?i)ninetieth`), "90th"},
}

var tenPrefixes = []struct {
	word string
	val  int
}{
	{"twenty", 20}, {"thirty", 30}, {"forty", 40}, {"fourty", 40}, {"fifty", 50},
	{"sixty", 60}, {"seventy", 70}, {"eighty", 80}, {"ninety", 90},
}

var bigPrefixes = []struct {
	word string
	val  int
}{
	{"hundred", 100}, {"thousand", 1000}, {"million", 1_000_000},
	{"billion", 1_000_000_000}, {"trillion", 1_000_000_000_000},
}

func gsubFunc(re *regexp.Regexp, s string, fn func([]string) string) string {
	return re.ReplaceAllStringFunc(s, func(m string) string {
		return fn(re.FindStringSubmatch(m))
	})
}

var reAndition = regexp.MustCompile(`(?i)<num>(\d+)( | and )<num>(\d+)`)

// andition merges adjacent <num> values ("<num>100 and <num>36" -> "<num>136"),
// porting Numerizer.andition's StringScanner loop.
func andition(s string) string {
	for {
		loc := reAndition.FindStringSubmatchIndex(s)
		if loc == nil {
			return s
		}
		// require the match to be followed by a non-word char or end of string.
		if end := loc[1]; end < len(s) && isWordByte(s[end]) {
			return s
		}
		g1 := s[loc[2]:loc[3]]
		sep := s[loc[4]:loc[5]]
		g3 := s[loc[6]:loc[7]]
		if strings.Contains(sep, "and") || len(g1) > len(g3) {
			s = s[:loc[0]] + "<num>" + strconv.Itoa(atoi(g1)+atoi(g3)) + s[loc[1]:]
			continue
		}
		return s
	}
}

func isWordByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
