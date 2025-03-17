package ai

import (
	"bytes"
	"text/template"

	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

type Prompt struct {
	reports map[string]*metrics.Report
	history *OptimizationSuggestion
	schema  any
	system  string
	user    string
}

type PromptOption func(*Prompt)

func NewPrompt(opts ...PromptOption) *Prompt {
	prompt := &Prompt{
		system: templates["system_prompt"],
		user:   templates["user_prompt"],
	}

	for _, opt := range opts {
		opt(prompt)
	}

	return prompt
}

func WithReport(name string, report *metrics.Report) PromptOption {
	return func(prompt *Prompt) {
		prompt.reports[name] = report

		tmpl, err := template.New("user_prompt").Parse(prompt.user)
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, report); err != nil {
			panic(err)
		}

		prompt.user = buf.String()
	}
}

func WithSchema(schema any) PromptOption {
	return func(prompt *Prompt) {
		prompt.schema = schema
	}
}

func WithHistory(history *OptimizationSuggestion) PromptOption {
	return func(prompt *Prompt) {
		prompt.history = history
	}
}

func WithTemplate(name string, tmpl string) PromptOption {
	return func(prompt *Prompt) {
		tmpl, err := template.New(name).Parse(tmpl)
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, prompt.reports); err != nil {
			panic(err)
		}

		templates[name] = buf.String()
	}
}

var templates = map[string]string{
	"system_prompt": `
	You are a MongoDB performance expert.
	
	You will be given a MongoDB database and a list of queries.
	
	You will need to analyze the queries and suggest optimizations.
	
	You will need to provide a detailed explanation of the queries and the optimizations.	
	`,
	"user_prompt": `
	Please analyze the following MongoDB metrics and suggest optimizations.

	{{.Report}}
	`,
	"measurement_prompt": `
	Below are two performance reports, plus a list of optimizations that have been applied.

	Please compare the before and after metrics to determine if the optimizations were beneficial.

	Then generate new optimizations, which either:

	1. Revert a previous optimization, based on it showing no improvement, or degradation
	2. Refine an existing optimization, based on it showing improvement
	3. Suggest a new optimization

	Before:

	{{ .Before }}

	Optimizations:

	{{ .Optimizations }}

	After:

	{{ .After }}
	`,
}
