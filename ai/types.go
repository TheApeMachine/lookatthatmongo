package ai

import (
	"time"

	"github.com/invopop/jsonschema"
)

/*
OptimizationSuggestion represents an AI-generated optimization recommendation.
It contains comprehensive information about a suggested MongoDB optimization,
including the problem, solution, and validation steps.
*/
type OptimizationSuggestion struct {
	Category     string   `json:"category" jsonschema:"enum=index,enum=query,enum=schema,enum=configuration" jsonschema_description:"The type of optimization being suggested"`
	Impact       string   `json:"impact" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Expected impact of implementing this suggestion"`
	Confidence   float64  `json:"confidence" jsonschema_description:"AI's confidence in this suggestion (0-1)"`
	Problem      Problem  `json:"problem" jsonschema_description:"Details about the identified issue"`
	Solution     Solution `json:"solution" jsonschema_description:"Proposed solution details"`
	Validation   []string `json:"validation" jsonschema_description:"Steps to validate the optimization's effectiveness"`
	RollbackPlan string   `json:"rollback_plan" jsonschema_description:"Steps to revert the changes if needed"`
}

/*
Problem describes the identified performance issue in the MongoDB database.
It includes a description, relevant metrics, when it was first detected, and its severity.
*/
type Problem struct {
	Description string    `json:"description" jsonschema_description:"Detailed description of the performance issue"`
	Metrics     []Metric  `json:"metrics" jsonschema_description:"Relevant metrics that indicate the problem"`
	FirstSeen   time.Time `json:"first_seen" jsonschema_description:"When this issue was first detected"`
	Severity    string    `json:"severity" jsonschema:"enum=critical,enum=high,enum=medium,enum=low" jsonschema_description:"How severe the problem is"`
}

/*
Solution contains the detailed optimization proposal to address the identified problem.
It includes a description, MongoDB commands to implement the solution, relevant resources,
and an estimate of the time and risk involved.
*/
type Solution struct {
	Description  string   `json:"description" jsonschema_description:"Detailed description of the proposed solution"`
	Commands     []string `json:"commands" jsonschema_description:"MongoDB commands to implement the solution"`
	Resources    []string `json:"resources" jsonschema_description:"Links to relevant documentation or resources"`
	TimeEstimate struct {
		Minutes int    `json:"minutes" jsonschema_description:"Estimated time to implement in minutes"`
		Risk    string `json:"risk" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Risk level of implementing this change"`
	} `json:"time_estimate" jsonschema_description:"Implementation time and risk estimation"`
}

/*
Metric represents a specific performance measurement related to the MongoDB database.
It includes the metric name, current value, unit, threshold for acceptable performance,
and the trend of the metric over time.
*/
type Metric struct {
	Name      string  `json:"name" jsonschema_description:"Name of the metric"`
	Value     float64 `json:"value" jsonschema_description:"Current value"`
	Unit      string  `json:"unit" jsonschema_description:"Unit of measurement"`
	Threshold float64 `json:"threshold" jsonschema_description:"Threshold for acceptable performance"`
	Trend     string  `json:"trend" jsonschema:"enum=improving,enum=stable,enum=degrading" jsonschema_description:"How this metric is trending"`
}

/*
GenerateSchema creates a JSON schema for structured outputs.
It uses reflection to generate a schema based on the provided type T.
*/
func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

var (
	OptimizationSuggestionSchema = GenerateSchema[OptimizationSuggestion]()
)
