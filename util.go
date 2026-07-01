// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "strconv"

// atoi parses a decimal integer, returning 0 for the empty/invalid string —
// matching Ruby's String#to_i behaviour the port relies on.
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
