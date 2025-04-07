package ai

import (
	"github.com/invopop/jsonschema"
)

/*
OptimizationSuggestion represents an AI-generated optimization recommendation.
It contains comprehensive information about a suggested MongoDB optimization,
including the problem, solution, and validation steps.
*/
type OptimizationSuggestion struct {
	Category   string   `json:"category" jsonschema:"enum=index,enum=query,enum=schema,enum=configuration" jsonschema_description:"The type of optimization being suggested"`
	Impact     string   `json:"impact" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Expected impact of implementing this suggestion"`
	Confidence float64  `json:"confidence" jsonschema_description:"AI's confidence in this suggestion (0-1)"`
	Problem    Problem  `json:"problem" jsonschema_description:"Details about the identified issue"`
	Solution   Solution `json:"solution" jsonschema_description:"Proposed solution details including specific operation parameters"`
	Validation []string `json:"validation" jsonschema_description:"Steps/metrics to validate the optimization's effectiveness"`
}

/*
Problem describes the identified performance issue in the MongoDB database.
It includes a description, relevant metrics, when it was first detected, and its severity.
*/
type Problem struct {
	Description string   `json:"description" jsonschema_description:"Detailed description of the performance issue"`
	Metrics     []Metric `json:"metrics" jsonschema_description:"Relevant metrics that indicate the problem"`
	FirstSeen   string   `json:"first_seen" jsonschema_description:"When this issue was first detected (ISO 8601 string format)"`
	Severity    string   `json:"severity" jsonschema:"enum=critical,enum=high,enum=medium,enum=low" jsonschema_description:"How severe the problem is"`
}

// IndexKey represents a field and its order in an index key specification.
// We use map[string]int for flexibility (e.g., {"fieldName": 1, "otherField": -1}).
type IndexKey map[string]int

// IndexOptions represents optional parameters for index creation.
type IndexOptions struct {
	Name               string `json:"name,omitempty" jsonschema_description:"Optional: Custom name for the index. Auto-generated if omitted."`
	Unique             bool   `json:"unique,omitempty" jsonschema_description:"Optional: If true, creates a unique index."`
	Sparse             bool   `json:"sparse,omitempty" jsonschema_description:"Optional: If true, creates a sparse index."`
	ExpireAfterSeconds *int   `json:"expireAfterSeconds,omitempty" jsonschema_description:"Optional: TTL index expiration time in seconds."`
	// Add other relevant options like background, partialFilterExpression, etc. as needed
}

// IndexOperation defines parameters for creating or dropping an index.
type IndexOperation struct {
	Action     string       `json:"action" jsonschema:"enum=createIndex,enum=dropIndex" jsonschema_description:"Action to perform: createIndex or dropIndex"`
	Collection string       `json:"collection" jsonschema_description:"The target collection name"`
	Keys       IndexKey     `json:"keys,omitempty" jsonschema_description:"Required for createIndex: The index key specification (e.g., {'fieldName': 1})"`
	Name       string       `json:"name,omitempty" jsonschema_description:"Required for dropIndex, optional for createIndex (if omitted, uses auto-generated name or options.name)"`
	Options    IndexOptions `json:"options,omitempty" jsonschema_description:"Optional parameters for createIndex"`
}

/*
Solution contains the detailed optimization proposal, now with structured operation details.
*/
type Solution struct {
	Description    string                 `json:"description" jsonschema_description:"Detailed description of the proposed solution"`
	Operations     []IndexOperation       `json:"operations" jsonschema_description:"Specific database operations needed (currently supports index operations)"`
	Resources      []string               `json:"resources,omitempty" jsonschema_description:"Optional: Links to relevant documentation or resources"`
	Implementation *ImplementationDetails `json:"implementation,omitempty" jsonschema_description:"Optional: Details about implementation complexity"`
}

// ImplementationDetails provides context on the solution's complexity.
type ImplementationDetails struct {
	EstimatedEffort string `json:"estimated_effort" jsonschema:"enum=low,enum=medium,enum=high" jsonschema_description:"Qualitative estimate of effort/time required"`
	RiskLevel       string `json:"risk_level" jsonschema:"enum=low,enum=medium,enum=high" jsonschema_description:"Potential risk level of implementing this change"`
}

/*
Metric represents a specific performance measurement related to the MongoDB database.
*/
type Metric struct {
	Name      string  `json:"name" jsonschema_description:"Name of the metric"`
	Value     float64 `json:"value" jsonschema_description:"Current value"`
	Unit      string  `json:"unit,omitempty" jsonschema_description:"Optional: Unit of measurement"`
	Threshold float64 `json:"threshold,omitempty" jsonschema_description:"Optional: Threshold for acceptable performance"`
	Trend     string  `json:"trend,omitempty" jsonschema:"enum=improving,enum=stable,enum=degrading" jsonschema_description:"Optional: How this metric is trending"`
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
	return reflector.Reflect(&v)
}

var (
	OptimizationSuggestionSchema = GenerateSchema[OptimizationSuggestion]()
)
