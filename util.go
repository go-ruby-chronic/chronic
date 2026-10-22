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

// atoi64 is atoi at 64 bits on every target. The numerizer handles values up
// to "trillion" (10^12), which a 32-bit int cannot hold; Ruby's Integer has no
// such limit, and int64 keeps 32-bit builds answering what 64-bit ones do.
func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
