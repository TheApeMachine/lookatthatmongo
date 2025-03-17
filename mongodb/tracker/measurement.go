package tracker

import (
	"context"

	"github.com/theapemachine/lookatthatmongo/ai"
)

/*
Measurement is a measurement of the performance of an optimization.
*/
type Measurement struct {
	conn    *ai.Conn
	history *History
}

type MeasurementOptionFn func(*Measurement)

func NewMeasurement(opts ...MeasurementOptionFn) *Measurement {
	opt := &Measurement{}

	for _, fn := range opts {
		fn(opt)
	}

	return opt
}

func WithConn(conn *ai.Conn) MeasurementOptionFn {
	return func(m *Measurement) {
		m.conn = conn
	}
}

func WithHistory(history *History) MeasurementOptionFn {
	return func(m *Measurement) {
		m.history = history
	}
}

func (measurement *Measurement) Measure(ctx context.Context) (*ai.OptimizationSuggestion, error) {
	prompt := ai.NewPrompt(
		ai.WithHistory(measurement.history.optimizations[0]),
		ai.WithReport("before", measurement.history.report),
		ai.WithReport("after", measurement.history.report),
		ai.WithSchema(ai.OptimizationSuggestionSchema),
	)

	response, err := measurement.conn.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	measurement.history.optimizations[0] = response.(*ai.OptimizationSuggestion)

	return response.(*ai.OptimizationSuggestion), nil
}
