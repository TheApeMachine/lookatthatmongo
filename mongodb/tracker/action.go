package tracker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"github.com/theapemachine/lookatthatmongo/mongodb/optimizer"
	"github.com/theapemachine/lookatthatmongo/storage"
)

// ActionType represents the type of action to take based on measurement results
type ActionType string

// String implements the Stringer interface for ActionType
func (a ActionType) String() string {
	return string(a)
}

const (
	// ActionNone indicates no action is needed
	ActionNone ActionType = "none"
	// ActionAlert indicates an alert should be sent
	ActionAlert ActionType = "alert"
	// ActionRollback indicates a rollback should be performed
	ActionRollback ActionType = "rollback"
	// ActionOptimize indicates further optimization should be performed
	ActionOptimize ActionType = "optimize"
)

// ActionResult represents the result of an action
type ActionResult struct {
	Type        ActionType `json:"type"`
	Success     bool       `json:"success"`
	Description string     `json:"description"`
	Timestamp   time.Time  `json:"timestamp"`
}

/*
ActionHandler manages actions based on optimization measurements.
It decides whether to keep, rollback, or modify optimizations based on their impact.
*/
type ActionHandler struct {
	storage   storage.Storage
	aiConn    *ai.Conn
	history   *History
	threshold float64 // Improvement threshold percentage
	optimizer optimizer.Optimizer
}

/*
ActionHandlerOptionFn is a function type for configuring an ActionHandler instance.
It follows the functional options pattern for flexible configuration.
*/
type ActionHandlerOptionFn func(*ActionHandler)

/*
NewActionHandler creates a new ActionHandler instance with the given options.
It initializes the action handler and applies any provided options.
*/
func NewActionHandler(opts ...ActionHandlerOptionFn) *ActionHandler {
	handler := &ActionHandler{
		threshold: 5.0, // Default improvement threshold
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

/*
WithStorage is an option function that sets the storage for the action handler.
*/
func WithStorage(storage storage.Storage) ActionHandlerOptionFn {
	return func(h *ActionHandler) {
		h.storage = storage
	}
}

/*
WithAIConn is an option function that sets the AI connection for the action handler.
*/
func WithAIConn(conn *ai.Conn) ActionHandlerOptionFn {
	return func(h *ActionHandler) {
		h.aiConn = conn
	}
}

/*
WithActionHistory is an option function that sets the optimization history.
*/
func WithActionHistory(history *History) ActionHandlerOptionFn {
	return func(h *ActionHandler) {
		h.history = history
	}
}

/*
WithThreshold is an option function that sets the improvement threshold.
*/
func WithThreshold(threshold float64) ActionHandlerOptionFn {
	return func(h *ActionHandler) {
		h.threshold = threshold
	}
}

/*
WithOptimizer is an option function that sets the optimizer.
*/
func WithOptimizer(optimizer optimizer.Optimizer) ActionHandlerOptionFn {
	return func(h *ActionHandler) {
		h.optimizer = optimizer
	}
}

// ProcessMeasurement processes a measurement and takes appropriate action
func (h *ActionHandler) ProcessMeasurement(ctx context.Context, suggestion *ai.OptimizationSuggestion, beforeReport, afterReport *metrics.Report) (*ActionResult, error) {
	logger.Info("Processing measurement results",
		"category", suggestion.Category,
		"confidence", suggestion.Confidence)

	// Create a prompt for the AI to determine the action to take
	prompt := ai.NewPrompt(
		ai.WithHistory(suggestion),
		ai.WithReport("before", beforeReport),
		ai.WithReport("after", afterReport),
	)

	// Ask the AI to determine what action to take
	response, err := h.aiConn.Generate(ctx, prompt)
	if err != nil {
		logger.Error("Failed to generate action recommendation", "error", err)
		return nil, fmt.Errorf("failed to generate action recommendation: %w", err)
	}

	actionSuggestion := response.(*ai.OptimizationSuggestion)

	// Determine action type based on AI suggestion
	var actionType ActionType
	improvement := getImprovementFromSuggestion(actionSuggestion)

	if improvement < 0 {
		actionType = ActionRollback
	} else if improvement < h.threshold {
		actionType = ActionAlert
	} else if actionSuggestion.Category == "optimize" {
		actionType = ActionOptimize
	} else {
		actionType = ActionNone
	}

	// Execute the determined action
	return h.ExecuteAction(ctx, actionType, actionSuggestion, improvement)
}

// ExecuteAction executes the determined action
func (h *ActionHandler) ExecuteAction(ctx context.Context, actionType ActionType, suggestion *ai.OptimizationSuggestion, improvement float64) (*ActionResult, error) {
	result := &ActionResult{
		Type:      actionType,
		Timestamp: time.Now(),
	}

	switch actionType {
	case ActionNone:
		result.Success = true
		result.Description = "No action required"
		logger.Info("No further action required",
			"improvement", improvement,
			"category", suggestion.Category)

	case ActionAlert:
		// In a real implementation, this would send an alert via email, Slack, etc.
		logger.Warn("Alert: Optimization improvement below threshold",
			"improvement", improvement,
			"threshold", h.threshold,
			"category", suggestion.Category)
		result.Success = true
		result.Description = fmt.Sprintf("Alert sent: Optimization improvement of %.2f%% below threshold of %.2f%%", improvement, h.threshold)

	case ActionRollback:
		// Trigger the rollback process using the optimizer
		if h.optimizer != nil {
			logger.Warn("Performing rollback due to performance degradation",
				"improvement", improvement,
				"category", suggestion.Category)

			if err := h.optimizer.Rollback(ctx, suggestion); err != nil {
				logger.Error("Rollback failed", "error", err)
				result.Success = false
				result.Description = fmt.Sprintf("Rollback failed: %v", err)
			} else {
				result.Success = true
				result.Description = fmt.Sprintf("Rollback completed successfully due to performance degradation of %.2f%%", improvement)
			}
		} else {
			logger.Error("Cannot perform rollback, optimizer not available")
			result.Success = false
			result.Description = "Rollback required but optimizer not available"
		}

	case ActionOptimize:
		// Apply additional optimizations using the optimizer
		if h.optimizer != nil {
			logger.Info("Applying additional optimizations",
				"category", suggestion.Category)

			if err := h.optimizer.Apply(ctx, suggestion); err != nil {
				logger.Error("Additional optimization failed", "error", err)
				result.Success = false
				result.Description = fmt.Sprintf("Additional optimization failed: %v", err)
			} else {
				result.Success = true
				result.Description = "Additional optimization applied successfully"
			}
		} else {
			logger.Error("Cannot perform additional optimization, optimizer not available")
			result.Success = false
			result.Description = "Additional optimization required but optimizer not available"
		}

	default:
		result.Success = false
		result.Description = fmt.Sprintf("Unknown action type: %s", actionType)
		// Change the Type to a string "unknown" for consistency in tests
		result.Type = ActionType("unknown")
	}

	// Store the action result in the optimization record
	if h.storage != nil {
		// Create a record of this optimization and action
		record := &storage.OptimizationRecord{
			Timestamp:        time.Now(),
			DatabaseName:     h.history.GetDatabaseName(),
			BeforeReport:     h.history.GetBeforeReport(),
			AfterReport:      h.history.GetAfterReport(),
			Suggestion:       suggestion,
			Applied:          true,
			Success:          result.Success,
			ImprovementPct:   improvement,
			RollbackRequired: actionType == ActionRollback,
			RollbackSuccess:  actionType == ActionRollback && result.Success,
		}

		if err := h.storage.SaveOptimizationRecord(ctx, record); err != nil {
			logger.Error("Failed to save optimization record", "error", err)
			// Continue execution even if storage fails
		}
	}

	return result, nil
}

// getImprovementFromSuggestion extracts the improvement percentage from the AI suggestion
func getImprovementFromSuggestion(suggestion *ai.OptimizationSuggestion) float64 {
	// Look for an improvement metric in the suggestion
	for _, metric := range suggestion.Problem.Metrics {
		if strings.Contains(strings.ToLower(metric.Name), "improvement") ||
			strings.Contains(strings.ToLower(metric.Name), "performance") {
			return metric.Value
		}
	}

	// Default improvement if not found
	return 0.0
}
