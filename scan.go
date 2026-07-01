// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "regexp"

// scanItem pairs a regexp with the tag type it yields, mirroring the Hash{Regexp
// => Symbol} the 0.10.2 scanners pass to scan_for. (0.10.2 matches only by
// Regexp against token.word — there is no positional String matching.)
type scanItem struct {
	re  *regexp.Regexp
	sym string
}

func item(re *regexp.Regexp, sym string) scanItem { return scanItem{re: re, sym: sym} }

// scanFor returns the sym of the first item whose regexp matches the word, and
// whether any matched.
func scanFor(word string, items []scanItem) (string, bool) {
	for _, it := range items {
		if it.re.MatchString(word) {
			return it.sym, true
		}
	}
	return "", false
}
