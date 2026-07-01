// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "strings"

// patElem is one element of a handler pattern. A pattern is a sequence of
// elements; each element is either a tag class (optionally optional) or a set of
// alternative tag classes, or a named sub-definition reference ("time", "anchor")
// mirroring the Ruby String elements.
type patElem struct {
	// alternatives is the set of tag-class options for this position. Matching
	// succeeds if the current token is-a any of them.
	alternatives []patTag
	// sub is a non-empty sub-definition name ("time", "anchor"), matched
	// recursively; when set, alternatives is ignored.
	sub      string
	optional bool // the element (or whole sub) is optional
}

type patTag struct {
	class    tagClass
	optional bool
}

// handler pairs a pattern with the method that builds a span from matched tokens.
type handler struct {
	pattern []patElem
	fn      handlerFn
}

// handlerFn builds a Span (or nil) from the good tokens, the parser, and options.
type handlerFn func(p *parser, tokens []*token, opts *options) *Span

// tagClassByName maps the snake_case symbol used in the gem's patterns to the
// tag class constant, so patterns can be written declaratively.
var tagClassByName = map[string]tagClass{
	"repeater":             tcRepeater,
	"repeater_time":        tcRepeaterTime,
	"repeater_day_portion": tcRepeaterDayPortion,
	"repeater_day_name":    tcRepeaterDayName,
	"repeater_month_name":  tcRepeaterMonthName,
	"repeater_month":       tcRepeaterMonth,
	"scalar":               tcScalar,
	"scalar_day":           tcScalarDay,
	"scalar_month":         tcScalarMonth,
	"scalar_year":          tcScalarYear,
	"ordinal":              tcOrdinal,
	"ordinal_day":          tcOrdinalDay,
	"grabber":              tcGrabber,
	"pointer":              tcPointer,
	"separator":            tcSeparator,
	"separator_slash":      tcSeparatorSlash,
	"separator_dash":       tcSeparatorDash,
	"separator_comma":      tcSeparatorComma,
	"separator_at":         tcSeparatorAt,
	"separator_in":         tcSeparatorIn,
	"separator_on":         tcSeparatorOn,
	"separator_and":        tcSeparatorAnd,
	"time_zone":            tcTimeZone,
}

// pt parses a compact spec string into a patElem. A trailing '?' marks optional.
// A '|'-joined spec is a set of alternatives (each may be optional). A spec that
// names a sub-definition ("time", "anchor") becomes a sub reference.
func pt(spec string) patElem {
	if spec == "time" || spec == "time?" || spec == "anchor" || spec == "anchor?" {
		opt := strings.HasSuffix(spec, "?")
		name := strings.TrimSuffix(spec, "?")
		return patElem{sub: name, optional: opt}
	}
	parts := strings.Split(spec, "|")
	e := patElem{}
	for _, p := range parts {
		opt := strings.HasSuffix(p, "?")
		name := strings.TrimSuffix(p, "?")
		e.alternatives = append(e.alternatives, patTag{class: tagClassByName[name], optional: opt})
		if opt {
			e.optional = true
		}
	}
	return e
}

// definitions holds the resolved grammar for the current endian preference.
type definitions struct {
	time   []handler
	date   []handler
	anchor []handler
	arrow  []handler
	narrow []handler
	endian []handler
}

// match ports Handler#match. It walks the pattern against tokens, honouring
// optional elements, alternative sets, and recursive sub-definitions.
func (h *handler) match(tokens []*token, defs *definitions) bool {
	tokenIndex := 0
	for _, elem := range h.pattern {
		if elem.sub != "" {
			// String element: optional && no tokens left => whole handler matches.
			if elem.optional && tokenIndex == len(tokens) {
				return true
			}
			subHandlers := defs.byName(elem.sub)
			for _, sh := range subHandlers {
				if sh.match(tokens[tokenIndex:], defs) {
					return true
				}
			}
			// A sub element that neither short-circuits nor matches leaves
			// tokenIndex unchanged; the loop moves on (mirroring Ruby which does
			// not advance for String elements).
			continue
		}
		// Symbol / alternative-set element.
		wasOptional := false
		matchedHere := false
		for i, alt := range elem.alternatives {
			if tokenIndex < len(tokens) && tokens[tokenIndex].getTag(alt.class) != nil {
				tokenIndex++
				matchedHere = true
				break
			}
			if alt.optional {
				wasOptional = true
				continue
			}
			if i+1 < len(elem.alternatives) {
				continue
			}
			if !wasOptional {
				return false
			}
		}
		_ = matchedHere
	}
	return tokenIndex == len(tokens)
}

func (d *definitions) byName(name string) []handler {
	switch name {
	case "time":
		return d.time
	case "anchor":
		return d.anchor
	case "date":
		return d.date
	}
	return nil
}
