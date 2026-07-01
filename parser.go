// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import "time"

// parser mirrors Chronic::Parser: it holds the mutable @now the handlers thread
// through the repeater pipeline (Parser#now= is reassigned inside day_or_time).
type parser struct {
	now time.Time
}

// parseToSpan runs the whole pipeline for text under opts and returns the
// matched Span or nil. It recovers construct's day-overflow panic as nil, the
// way the gem's rescue does.
func parseToSpan(text string, opts *options) (result *Span) {
	defer func() {
		if r := recover(); r != nil {
			if r == errDayOverflow {
				result = nil
				return
			}
			panic(r)
		}
	}()
	opts.text = text
	p := &parser{now: opts.now}
	tokens := tokenizePipeline(text, opts)
	return p.tokensToSpan(tokens, opts)
}

// tokenizePipeline ports Parser#tokenize: pre-normalise, tokenise, run every
// scanner, then keep only tagged tokens.
func tokenizePipeline(text string, opts *options) []*token {
	text = preNormalize(text, opts.ambiguousYearFutureBias, opts.now.Year())
	tokens := tokenize(text)

	scanRepeater(tokens, opts)
	scanGrabber(tokens)
	scanPointer(tokens)
	scanScalar(tokens, opts)
	scanOrdinal(tokens, opts)
	scanSeparator(tokens)
	scanSign(tokens)
	scanTimeZone(tokens)

	out := tokens[:0]
	for _, t := range tokens {
		if t.tagged() {
			out = append(out, t)
		}
	}
	return out
}

// tokensToSpan ports Parser#tokens_to_span: try endian+date, then anchor, arrow,
// narrow definition groups in order.
func (p *parser) tokensToSpan(tokens []*token, opts *options) *Span {
	defs := buildDefinitions(opts)

	for _, h := range append(append([]handler{}, defs.endian...), defs.date...) {
		if h.match(tokens, defs) {
			good := filterNot(tokens, tcSeparator)
			return h.fn(p, good, opts)
		}
	}

	for _, h := range defs.anchor {
		if h.match(tokens, defs) {
			good := filterNot(tokens, tcSeparator)
			return h.fn(p, good, opts)
		}
	}

	for _, h := range defs.arrow {
		if h.match(tokens, defs) {
			good := filterOutArrowSeparators(tokens)
			return h.fn(p, good, opts)
		}
	}

	for _, h := range defs.narrow {
		if h.match(tokens, defs) {
			return h.fn(p, tokens, opts)
		}
	}
	return nil
}

// filterNot drops tokens carrying a tag of the given class.
func filterNot(tokens []*token, c tagClass) []*token {
	out := make([]*token, 0, len(tokens))
	for _, t := range tokens {
		if t.getTag(c) == nil {
			out = append(out, t)
		}
	}
	return out
}

// filterOutArrowSeparators drops the separators the arrow group ignores.
func filterOutArrowSeparators(tokens []*token) []*token {
	out := make([]*token, 0, len(tokens))
	for _, t := range tokens {
		if t.getTag(tcSeparatorAt) != nil || t.getTag(tcSeparatorSlash) != nil ||
			t.getTag(tcSeparatorDash) != nil || t.getTag(tcSeparatorComma) != nil ||
			t.getTag(tcSeparatorAnd) != nil {
			continue
		}
		out = append(out, t)
	}
	return out
}
