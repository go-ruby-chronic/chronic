# frozen_string_literal: true
#
# Basic usage of Chronic — natural-language date/time parsing — from Ruby.
# Runs under go-embedded-ruby (rbgo); see examples/README.md.

require "chronic"

# Anchor "now" so the output is reproducible (2026-07-10 12:00:00 UTC).
# In real code you would just call Chronic.parse("tomorrow") and let it use
# the current time.
now = Time.at(1_783_684_800)

# A phrase resolves to a single Time (Chronic guesses a point in the span).
tomorrow = Chronic.parse("tomorrow", now: now)
puts tomorrow.strftime("%Y-%m-%d %H:%M:%S")     # => 2026-07-11 12:00:00

# Relative and calendar phrases both work.
puts Chronic.parse("next tuesday", now: now).strftime("%Y-%m-%d")  # => 2026-07-14
puts Chronic.parse("3 days ago", now: now).strftime("%Y-%m-%d")    # => 2026-07-07

# An unparseable phrase returns nil rather than raising.
p Chronic.parse("total gibberish", now: now)     # => nil

# guess: false returns the whole span as a [begin, end] pair of Times.
first, last = Chronic.parse("this month", now: now, guess: false)
puts "#{first.strftime('%Y-%m-%d')}..#{last.strftime('%Y-%m-%d')}"  # => 2026-07-01..2026-08-01

# context: :past disambiguates a bare date toward the previous occurrence.
puts Chronic.parse("may 27", now: now, context: :past).strftime("%Y-%m-%d")  # => 2026-05-27
