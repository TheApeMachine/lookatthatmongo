package ai

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// defaultTemplates contains the default templates used by the system.
// This is a constant map, so it's safe to be package-level.
var defaultTemplates = map[string]string{
	"system_prompt": `
	You are a MongoDB performance expert.
	You MUST generate responses STRICTLY conforming to the provided JSON schema.
	
	CRITICAL REQUIREMENTS:
	- Analyze the provided MongoDB metrics and suggest optimizations.
	- Focus on ONE specific optimization category (index, query, etc.) per suggestion.
	- For the 'solution.operations' array, provide the specific parameters needed to perform the action.
	  - For 'createIndex', specify 'collection', 'keys', and optionally 'options.name' or other options.
	  - For 'dropIndex', specify 'collection' and 'name'.
	- DO NOT provide raw MongoDB commands or shell syntax.
	- Provide a detailed explanation for the problem and the reasoning behind your suggested solution.
	- Ensure the entire output is a single JSON object matching the schema.
	`,
	"user_prompt": `
	Please analyze the following MongoDB metrics and suggest optimizations based on the provided schema.

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

// Prompt represents a prompt to be sent to the AI model.
type Prompt struct {
	reports   map[string]*metrics.Report
	history   *OptimizationSuggestion
	schema    any
	system    string
	user      string
	templates map[string]*template.Template // Pre-parsed templates
	tmplCache *sync.Map                     // Thread-safe template cache
	tmplMutex sync.RWMutex                  // Mutex for template operations
}

// PromptOption is a function type for configuring a Prompt instance.
type PromptOption func(*Prompt) error

// NewPrompt creates a new Prompt instance with the given options.
func NewPrompt(opts ...PromptOption) (*Prompt, error) {
	p := &Prompt{
		reports:   make(map[string]*metrics.Report),
		templates: make(map[string]*template.Template),
		tmplCache: &sync.Map{},
		system:    defaultTemplates["system_prompt"],
		user:      defaultTemplates["user_prompt"],
	}

	// Pre-parse default templates
	if err := p.parseDefaultTemplates(); err != nil {
		return nil, fmt.Errorf("failed to parse default templates: %w", err)
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return p, nil
}

// parseDefaultTemplates parses the default templates and stores them in the cache
func (p *Prompt) parseDefaultTemplates() error {
	for name, text := range defaultTemplates {
		tmpl, err := template.New(name).Parse(text)
		if err != nil {
			return fmt.Errorf("failed to parse default template %s: %w", name, err)
		}
		p.templates[name] = tmpl
	}
	return nil
}

// getTemplate returns a parsed template from cache or parses and caches it
func (p *Prompt) getTemplate(name, text string) (*template.Template, error) {
	// Check instance cache first
	p.tmplMutex.RLock()
	if tmpl, ok := p.templates[name]; ok {
		p.tmplMutex.RUnlock()
		return tmpl, nil
	}
	p.tmplMutex.RUnlock()

	// Check shared cache
	if cached, ok := p.tmplCache.Load(name); ok {
		if tmpl, ok := cached.(*template.Template); ok {
			return tmpl, nil
		}
	}

	// Parse and cache the template
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	p.tmplMutex.Lock()
	p.templates[name] = tmpl
	p.tmplMutex.Unlock()
	p.tmplCache.Store(name, tmpl)

	return tmpl, nil
}

// WithReport adds a metrics report to the prompt and renders the user prompt template.
func WithReport(name string, report *metrics.Report) PromptOption {
	return func(p *Prompt) error {
		p.reports[name] = report

		tmpl, err := p.getTemplate("user_prompt", p.user)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		var buf bytes.Buffer
		// Wrap the report in a map for template execution
		data := map[string]any{
			"Report": report,
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		p.user = buf.String()
		return nil
	}
}

// WithSchema sets the JSON schema for the prompt.
func WithSchema(schema any) PromptOption {
	return func(p *Prompt) error {
		p.schema = schema
		return nil
	}
}

// WithHistory adds optimization history to the prompt.
func WithHistory(history *OptimizationSuggestion) PromptOption {
	return func(p *Prompt) error {
		p.history = history
		return nil
	}
}

// WithTemplate sets a custom template for the prompt.
func WithTemplate(name string, text string) PromptOption {
	return func(p *Prompt) error {
		// Parse and validate the template
		tmpl, err := p.getTemplate(name, text)
		if err != nil {
			return fmt.Errorf("invalid template %s: %w", name, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, p.reports); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", name, err)
		}

		// Store the executed template result
		switch name {
		case "system_prompt":
			p.system = buf.String()
		case "user_prompt":
			p.user = buf.String()
		}

		return nil
	}
}
