# chronic — go-ruby-chronic

[![Go Reference](https://pkg.go.dev/badge/github.com/go-ruby-chronic/chronic.svg)](https://pkg.go.dev/github.com/go-ruby-chronic/chronic)
[![License: BSD-3-Clause](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/go-ruby-chronic/chronic/actions/workflows/ci.yml/badge.svg)](https://github.com/go-ruby-chronic/chronic/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby's
[`chronic`](https://github.com/mojombo/chronic) gem** — a natural-language date
and time parser. It turns English phrases such as `"tomorrow"`, `"next tuesday"`,
`"3 days ago"`, `"may 27th"` or `"2016-05-27"` into a concrete instant (or a
`Span`), faithfully mirroring the deterministic grammar of **chronic 0.10.2** —
**without any Ruby runtime**.

It is the natural-language time backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module with no dependency on the Ruby runtime — a
sibling of [go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) (the Onigmo
engine) and [go-ruby-erb](https://github.com/go-ruby-erb/erb) (the ERB compiler).

## Deterministic by construction

`Parse` is deterministic: pass a fixed anchor via `Options.Now` and the result
depends only on the input string, so it needs no interpreter and no wall clock.
Its output is differential-tested **byte-for-byte** against MRI's own `chronic`
gem.

## Usage

```go
import (
	"time"

	"github.com/go-ruby-chronic/chronic"
)

now := time.Date(2006, 8, 16, 14, 0, 0, 0, time.Local)
t, _ := chronic.Parse("tomorrow", chronic.Options{Now: now})
// t == 2006-08-17 12:00:00
```

## Tests & coverage

`go test ./...` runs the unit, golden and differential-oracle suites. The CI
gate enforces **100% statement coverage** and builds/tests on all six 64-bit Go
targets — `amd64`, `arm64`, `riscv64`, `loong64`, `ppc64le`, `s390x`.

## License

BSD-3-Clause. Copyright (c) the go-ruby-chronic/chronic authors.
