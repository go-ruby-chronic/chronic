// Copyright (c) the go-ruby-chronic/chronic authors
//
// SPDX-License-Identifier: BSD-3-Clause

package chronic

// buildDefinitions ports Parser#definitions for chronic 0.10.2. The endian block
// is reversed when little-endian precedence is requested.
func buildDefinitions(opts *options) *definitions {
	d := &definitions{}

	d.time = []handler{
		{[]patElem{pt("repeater_time"), pt("repeater_day_portion?")}, nil},
	}

	d.date = []handler{
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("scalar_day"), pt("repeater_time"), pt("separator_slash?|separator_dash?"), pt("time_zone"), pt("scalar_year")}, handleGeneric},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("scalar_day")}, handleRdnRmnSd},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("scalar_day"), pt("scalar_year")}, handleRdnRmnSdSy},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("ordinal_day")}, handleRdnRmnOd},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("ordinal_day"), pt("scalar_year")}, handleRdnRmnOdSy},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("scalar_day"), pt("separator_at?"), pt("time?")}, handleRdnRmnSd},
		{[]patElem{pt("repeater_day_name"), pt("repeater_month_name"), pt("ordinal_day"), pt("separator_at?"), pt("time?")}, handleRdnRmnOd},
		{[]patElem{pt("repeater_day_name"), pt("ordinal_day"), pt("separator_at?"), pt("time?")}, handleRdnOd},
		{[]patElem{pt("scalar_year"), pt("separator_slash|separator_dash"), pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_day"), pt("repeater_time"), pt("time_zone")}, handleGeneric},
		{[]patElem{pt("ordinal_day")}, handleGeneric},
		{[]patElem{pt("repeater_month_name"), pt("scalar_day"), pt("scalar_year")}, handleRmnSdSy},
		{[]patElem{pt("repeater_month_name"), pt("ordinal_day"), pt("scalar_year")}, handleRmnOdSy},
		{[]patElem{pt("repeater_month_name"), pt("scalar_day"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleRmnSdSy},
		{[]patElem{pt("repeater_month_name"), pt("ordinal_day"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleRmnOdSy},
		{[]patElem{pt("repeater_month_name"), pt("separator_slash?|separator_dash?"), pt("scalar_day"), pt("separator_at?"), pt("time?")}, handleRmnSd},
		{[]patElem{pt("repeater_time"), pt("repeater_day_portion?"), pt("separator_on?"), pt("repeater_month_name"), pt("scalar_day")}, handleRmnSdOn},
		{[]patElem{pt("repeater_month_name"), pt("ordinal_day"), pt("separator_at?"), pt("time?")}, handleRmnOd},
		{[]patElem{pt("ordinal_day"), pt("repeater_month_name"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleOdRmnSy},
		{[]patElem{pt("ordinal_day"), pt("repeater_month_name"), pt("separator_at?"), pt("time?")}, handleOdRmn},
		{[]patElem{pt("ordinal_day"), pt("grabber?"), pt("repeater_month"), pt("separator_at?"), pt("time?")}, handleOdRm},
		{[]patElem{pt("scalar_year"), pt("repeater_month_name"), pt("ordinal_day")}, handleSyRmnOd},
		{[]patElem{pt("repeater_time"), pt("repeater_day_portion?"), pt("separator_on?"), pt("repeater_month_name"), pt("ordinal_day")}, handleRmnOdOn},
		{[]patElem{pt("repeater_month_name"), pt("scalar_year")}, handleRmnSy},
		{[]patElem{pt("scalar_day"), pt("repeater_month_name"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleSdRmnSy},
		{[]patElem{pt("scalar_day"), pt("separator_slash?|separator_dash?"), pt("repeater_month_name"), pt("separator_at?"), pt("time?")}, handleSdRmn},
		{[]patElem{pt("scalar_year"), pt("separator_slash|separator_dash"), pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_day"), pt("separator_at?"), pt("time?")}, handleSySmSd},
		{[]patElem{pt("scalar_year"), pt("separator_slash|separator_dash"), pt("scalar_month")}, handleSySm},
		{[]patElem{pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_year")}, handleSmSy},
		{[]patElem{pt("scalar_day"), pt("separator_slash|separator_dash"), pt("repeater_month_name"), pt("separator_slash|separator_dash"), pt("scalar_year"), pt("repeater_time?")}, handleSmRmnSy},
		{[]patElem{pt("scalar_year"), pt("separator_slash|separator_dash"), pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar?"), pt("time_zone")}, handleGeneric},
	}

	d.anchor = []handler{
		{[]patElem{pt("separator_on?"), pt("grabber?"), pt("repeater"), pt("separator_at?"), pt("repeater?"), pt("repeater?")}, handleR},
		{[]patElem{pt("grabber?"), pt("repeater"), pt("repeater"), pt("separator?"), pt("repeater?"), pt("repeater?")}, handleR},
		{[]patElem{pt("repeater"), pt("grabber"), pt("repeater")}, handleRGR},
	}

	d.arrow = []handler{
		{[]patElem{pt("scalar"), pt("repeater"), pt("pointer")}, handleSRP},
		{[]patElem{pt("scalar"), pt("repeater"), pt("separator_and?"), pt("scalar"), pt("repeater"), pt("pointer"), pt("separator_at?"), pt("anchor")}, handleSRASRPA},
		{[]patElem{pt("pointer"), pt("scalar"), pt("repeater")}, handlePSR},
		{[]patElem{pt("scalar"), pt("repeater"), pt("pointer"), pt("separator_at?"), pt("anchor")}, handleSRPA},
	}

	d.narrow = []handler{
		{[]patElem{pt("ordinal"), pt("repeater"), pt("separator_in"), pt("repeater")}, handleORSR},
		{[]patElem{pt("ordinal"), pt("repeater"), pt("grabber"), pt("repeater")}, handleORGR},
	}

	endian := []handler{
		{[]patElem{pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_day"), pt("separator_slash|separator_dash"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleSmSdSy},
		{[]patElem{pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_day"), pt("separator_at?"), pt("time?")}, handleSmSd},
		{[]patElem{pt("scalar_day"), pt("separator_slash|separator_dash"), pt("scalar_month"), pt("separator_at?"), pt("time?")}, handleSdSm},
		{[]patElem{pt("scalar_day"), pt("separator_slash|separator_dash"), pt("scalar_month"), pt("separator_slash|separator_dash"), pt("scalar_year"), pt("separator_at?"), pt("time?")}, handleSdSmSy},
	}
	if opts.endianLittle {
		for i, j := 0, len(endian)-1; i < j; i, j = i+1, j-1 {
			endian[i], endian[j] = endian[j], endian[i]
		}
	}
	d.endian = endian

	return d
}
