// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"strconv"
	"strings"
	"time"
)

// parseGenericTimestamp ports handle_generic's Time.parse call for the ISO-8601
// / RFC-ish timestamps Chronic funnels through the :handle_generic definitions
// (e.g. "2012-08-02T13:00:00", "...+01:00", "...Z", "2012-01-03 01:00:00.100").
// It works on the ORIGINAL text (options[:text]), not the normalised form.
//
// A timestamp with an explicit offset (or Z) is normalised to the local zone,
// mirroring Ruby's Time.parse which returns an instant; a naive timestamp is
// interpreted in the local zone.
func parseGenericTimestamp(text string) (time.Time, bool) {
	s := strings.TrimSpace(text)
	// Normalise a space separator to 'T' so a single set of layouts suffices.
	if len(s) >= 11 && s[10] == ' ' {
		s = s[:10] + "T" + s[11:]
	}

	// Layouts with a zone (offset or Z).
	zoned := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999999Z07:00",
	}
	for _, l := range zoned {
		if t, err := time.Parse(l, s); err == nil {
			return t.In(TimeLoc), true
		}
	}
	// Naive layouts (no zone) parsed in local time.
	naive := []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15",
		"2006-01-02",
	}
	for _, l := range naive {
		if t, err := time.ParseInLocation(l, s, TimeLoc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// The ordinal-day-only "handle_generic" case ("ordinal_day" / "28th") and the
// day-name+time zone forms also route to handle_generic; Ruby there calls
// Time.parse on the raw text. We special-case the bare-ordinal form because the
// normalised text ("28th") is not an ISO timestamp: Ruby's Time.parse fills the
// missing parts from Time.now. That is inherently non-deterministic, so callers
// wanting determinism should avoid the bare-ordinal form; we still support it
// by combining the ordinal with the anchor is not possible here (no now), so we
// return false and let Parse yield nil for that corner. See README "deferred".
var _ = strconv.Atoi
