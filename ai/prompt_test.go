package ai

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// mockMonitor implements metrics.Monitor interface for testing
type mockMonitor struct{}

func (m *mockMonitor) GetServerStats(ctx any) (*metrics.ServerStats, error) { return nil, nil }
func (m *mockMonitor) GetDatabaseStats(ctx any, dbName string) (*metrics.DatabaseStats, error) {
	return nil, nil
}
func (m *mockMonitor) GetCollectionStats(ctx any, dbName, collName string) (*metrics.CollectionStats, error) {
	return nil, nil
}
func (m *mockMonitor) GetIndexStats(ctx any, dbName, collName string) ([]metrics.IndexStats, error) {
	return nil, nil
}

func newTestReport() *metrics.Report {
	return metrics.NewReport(&mockMonitor{})
}

func TestNewPrompt(t *testing.T) {
	prompt, err := NewPrompt()
	if err != nil {
		t.Fatalf("NewPrompt() failed: %v", err)
	}

	if prompt.system != defaultTemplates["system_prompt"] {
		t.Error("system prompt not initialized with default template")
	}

	if prompt.user != defaultTemplates["user_prompt"] {
		t.Error("user prompt not initialized with default template")
	}

	if prompt.reports == nil {
		t.Error("reports map not initialized")
	}

	if prompt.templates == nil {
		t.Error("templates map not initialized")
	}
}

func TestWithReport(t *testing.T) {
	report := newTestReport()
	report.Timestamp = time.Now()

	prompt, err := NewPrompt(WithReport("test", report))
	if err != nil {
		t.Fatalf("NewPrompt() with WithReport failed: %v", err)
	}

	if r, exists := prompt.reports["test"]; !exists || r != report {
		t.Error("report not properly stored in prompt")
	}

	// Since we're using a mock monitor, we'll just verify the timestamp was included
	if !strings.Contains(prompt.user, report.Timestamp.Format(time.RFC3339)) {
		t.Error("user prompt not updated with report timestamp")
	}
}

func TestWithTemplate(t *testing.T) {
	customTemplate := "Custom template with {{.Timestamp}}"
	prompt, err := NewPrompt(WithTemplate("user_prompt", customTemplate))
	if err != nil {
		t.Fatalf("NewPrompt() with WithTemplate failed: %v", err)
	}

	if prompt.templates["user_prompt"] == nil {
		t.Error("template should have been parsed and stored")
	}

	// Add a report and verify template execution
	report := newTestReport()
	report.Timestamp = time.Now()
	err = WithReport("test", report)(prompt)
	if err != nil {
		t.Fatalf("WithReport failed: %v", err)
	}

	if !strings.Contains(prompt.user, report.Timestamp.Format(time.RFC3339)) {
		t.Error("custom template not properly executed")
	}
}

func TestTemplateCacheConcurrency(t *testing.T) {
	prompt, err := NewPrompt()
	if err != nil {
		t.Fatalf("NewPrompt() failed: %v", err)
	}

	const concurrentAccess = 100
	var wg sync.WaitGroup
	wg.Add(concurrentAccess)

	for i := 0; i < concurrentAccess; i++ {
		go func() {
			defer wg.Done()
			tmpl, err := prompt.getTemplate("test_template", "Test {{.}} Concurrent")
			if err != nil {
				t.Errorf("getTemplate failed: %v", err)
				return
			}
			if tmpl == nil {
				t.Error("template should not be nil")
				return
			}
		}()
	}

	wg.Wait()
}

func TestWithSchemaAndHistory(t *testing.T) {
	schema := map[string]interface{}{"test": "schema"}
	history := &OptimizationSuggestion{
		Category:   "index",
		Impact:     "high",
		Confidence: 0.95,
		Problem: Problem{
			Description: "test problem",
			Severity:    "high",
		},
		Solution: Solution{
			Description: "test solution",
			Operations: []IndexOperation{
				{
					Action:     "createIndex",
					Collection: "testCollection",
					Keys:       IndexKey{"field": 1},
					Name:       "testIndexName",
				},
			},
		},
		Validation: []string{"test validation"},
	}

	prompt, err := NewPrompt(
		WithSchema(schema),
		WithHistory(history),
	)
	if err != nil {
		t.Fatalf("NewPrompt() with WithSchema and WithHistory failed: %v", err)
	}

	// For maps, we need to check if both are nil or both are not nil
	if (prompt.schema == nil) != (schema == nil) {
		t.Error("schema not properly stored")
	}

	if prompt.history != history {
		t.Error("history not properly stored")
	}
}

func TestInvalidTemplate(t *testing.T) {
	invalidTemplate := "{{.Invalid}"
	_, err := NewPrompt(WithTemplate("test", invalidTemplate))
	if err == nil {
		t.Error("expected error for invalid template, got nil")
	}
}

func TestTemplateExecution(t *testing.T) {
	template := "Report from: {{.Timestamp}}"
	report := newTestReport()
	report.Timestamp = time.Now()

	prompt, err := NewPrompt(
		WithTemplate("user_prompt", template),
		WithReport("test", report),
	)
	if err != nil {
		t.Fatalf("NewPrompt() failed: %v", err)
	}

	expected := "Report from: " + report.Timestamp.Format(time.RFC3339)
	if !strings.Contains(prompt.user, report.Timestamp.Format(time.RFC3339)) {
		t.Errorf("expected prompt to contain timestamp %q", expected)
	}
}
