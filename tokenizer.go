// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "strings"

// tokenize splits text into Tokens on runs of spaces, matching chronic 0.10.2's
// `text.split(' ')` (Ruby's split on a single space collapses whitespace and
// discards leading/trailing empties). pre_normalize has already inserted spaces
// around separators, so slashes/dashes/commas surface as their own tokens.
func tokenize(text string) []*token {
	var tokens []*token
	for _, w := range strings.Fields(text) {
		tokens = append(tokens, newToken(w))
	}
	return tokens
}
