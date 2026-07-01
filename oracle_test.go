// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The oracle tests run the real Ruby `chronic` gem 0.10.2 as a subprocess and
// assert that every phrase in the golden corpus resolves in Go exactly as MRI
// resolves it. They are version-gated: MRI < 4.0 (or a Ruby without the gem, or
// Windows, where ruby is absent on the CI lane) is skipped, so the deterministic
// golden-vector tests remain the sole coverage gate everywhere. The oracle keeps
// the golden.json corpus honest wherever ruby 4.0 + chronic 0.10.2 are present.

// oracleScript drives the gem over a batch of phrase/option requests read as
// JSON on stdin and emits one JSON result per line on stdout, anchored to the
// fixed goldenAnchor so results are deterministic. It also prints RUBY_VERSION
// and whether the gem loaded on the first two lines.
const oracleScript = `
begin
  require "chronic"
rescue LoadError
  STDOUT.puts RUBY_VERSION
  STDOUT.puts "no-gem"
  exit 0
end
STDOUT.puts RUBY_VERSION
STDOUT.puts "ok"
require "json"
now = Time.local(2006, 8, 16, 14, 0, 0)
reqs = JSON.parse(STDIN.read)
fmt = "%Y-%m-%d %H:%M:%S"
reqs.each do |r|
  opts = { now: now }
  o = r["opts"] || {}
  opts[:context] = o["context"].to_sym if o["context"]
  opts[:endian_precedence] = o["endian_precedence"].map(&:to_sym) if o["endian_precedence"]
  opts[:ambiguous_time_range] = o["ambiguous_time_range"].to_sym if o["ambiguous_time_range"] == "none"
  guess = (r["guess"] || "").to_s
  begin
    span = Chronic.parse(r["phrase"], opts.merge(guess: false))
    if span.nil?
      STDOUT.puts JSON.generate({"span" => nil, "guess" => nil})
      next
    end
    g =
      case guess
      when "begin" then Chronic.parse(r["phrase"], opts.merge(guess: :begin))
      when "end"   then Chronic.parse(r["phrase"], opts.merge(guess: :end))
      else Chronic.parse(r["phrase"], opts.merge(guess: :middle))
      end
    STDOUT.puts JSON.generate({
      "begin" => span.begin.strftime(fmt),
      "end"   => span.end.strftime(fmt),
      "guess" => g.nil? ? nil : g.strftime(fmt),
    })
  rescue => e
    STDOUT.puts JSON.generate({"error" => e.message})
  end
end
`

type oracleResult struct {
	Begin string  `json:"begin"`
	End   string  `json:"end"`
	Guess *string `json:"guess"`
	Span  *bool   `json:"span"` // present+nil signals "no parse"
	Error string  `json:"error"`
}

// versionAtLeast reports whether the dotted version string v is >= major.minor.
func versionAtLeast(v string, major, minor int) bool {
	parts := strings.SplitN(strings.TrimSpace(v), ".", 3)
	if len(parts) < 2 {
		return false
	}
	maj := atoi(parts[0])
	min := atoi(parts[1])
	if maj != major {
		return maj > major
	}
	return min >= minor
}

// runOracle launches ruby with oracleScript, feeding it the requests as JSON. It
// returns (results, true) when the gem ran, or (nil, false) with a skip reason
// when ruby/the gem is unavailable or too old.
func runOracle(t *testing.T, reqs []map[string]any) ([]oracleResult, bool) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("oracle: ruby absent on windows lane")
	}
	rubyBin, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("oracle: ruby not found")
	}
	in, err := json.Marshal(reqs)
	if err != nil {
		t.Fatalf("marshal requests: %v", err)
	}
	cmd := exec.Command(rubyBin, "-e", oracleScript)
	cmd.Env = append(cmd.Environ(), "TZ=UTC")
	cmd.Stdin = bytes.NewReader(in)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		t.Skipf("oracle: ruby run failed (%v): %s", err, errb.String())
	}
	sc := bufio.NewScanner(&out)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	if !sc.Scan() {
		t.Skip("oracle: no ruby version line")
	}
	ver := sc.Text()
	if !versionAtLeast(ver, 4, 0) {
		t.Skipf("oracle: ruby %s < 4.0", ver)
	}
	if !sc.Scan() {
		t.Skip("oracle: no gem status line")
	}
	if sc.Text() != "ok" {
		t.Skip("oracle: chronic gem not installed")
	}
	var results []oracleResult
	for sc.Scan() {
		var r oracleResult
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			t.Fatalf("oracle line %q: %v", sc.Text(), err)
		}
		results = append(results, r)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan oracle output: %v", err)
	}
	return results, true
}

// TestOracleGolden replays every golden phrase through both MRI and Go and
// asserts the span and the guessed instant agree, keeping golden.json faithful
// to chronic 0.10.2 wherever ruby 4.0 + the gem are present.
func TestOracleGolden(t *testing.T) {
	old := TimeLoc
	TimeLoc = time.UTC
	defer func() { TimeLoc = old }()

	vecs := loadGolden(t)
	reqs := make([]map[string]any, 0, len(vecs))
	for _, v := range vecs {
		opts := map[string]any{}
		if v.Opts.Context != "" {
			opts["context"] = v.Opts.Context
		}
		if len(v.Opts.EndianPrecedence) > 0 {
			opts["endian_precedence"] = v.Opts.EndianPrecedence
		}
		if s, ok := v.Opts.AmbiguousTimeRange.(string); ok && s == "none" {
			opts["ambiguous_time_range"] = "none"
		}
		guess := ""
		if gv, ok := v.Opts.Guess.(string); ok {
			guess = gv
		}
		reqs = append(reqs, map[string]any{"phrase": v.Phrase, "opts": opts, "guess": guess})
	}

	results, ran := runOracle(t, reqs)
	if !ran {
		return
	}
	if len(results) != len(vecs) {
		t.Fatalf("oracle returned %d results, want %d", len(results), len(vecs))
	}

	for i, v := range vecs {
		res := results[i]
		name := v.Phrase + optSuffix(v.Opts)
		t.Run(name, func(t *testing.T) {
			if res.Error != "" {
				t.Fatalf("gem raised on %q: %s", v.Phrase, res.Error)
			}
			o := v.Opts.toOptions()

			// Span comparison (guess disabled).
			so := o
			so.Guess = GuessNone
			gotSpan, ok := ParseSpan(v.Phrase, so)
			if res.Begin == "" { // gem parsed nothing
				if ok {
					t.Fatalf("Go parsed %q to %v, gem got nil", v.Phrase, gotSpan)
				}
				return
			}
			if !ok {
				t.Fatalf("Go got nil for %q, gem got %s..%s", v.Phrase, res.Begin, res.End)
			}
			if got := gotSpan.Begin.Format(goldenLayout); got != res.Begin {
				t.Errorf("span begin %q: Go %s, gem %s", v.Phrase, got, res.Begin)
			}
			if got := gotSpan.End.Format(goldenLayout); got != res.End {
				t.Errorf("span end %q: Go %s, gem %s", v.Phrase, got, res.End)
			}

			// Guessed instant.
			if res.Guess != nil {
				gotAt, ok := Parse(v.Phrase, o)
				if !ok {
					t.Fatalf("Go guess nil for %q, gem %s", v.Phrase, *res.Guess)
				}
				if got := gotAt.Format(goldenLayout); got != *res.Guess {
					t.Errorf("guess %q: Go %s, gem %s", v.Phrase, got, *res.Guess)
				}
			}
		})
	}
}
