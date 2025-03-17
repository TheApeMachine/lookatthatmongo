package tracker

import (
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

/*
History tracks the report and optimizations that relate to each other,
which will be used to validate (or invalidate) the benefits of the optimizations.
*/
type History struct {
	report        *metrics.Report
	optimizations []*ai.OptimizationSuggestion
}

type HistoryOptionFn func(*History)

func NewHistory(opts ...HistoryOptionFn) *History {
	opt := &History{}

	for _, fn := range opts {
		fn(opt)
	}

	return opt
}

func WithHistoryReport(report *metrics.Report) HistoryOptionFn {
	return func(h *History) {
		h.report = report
	}
}
