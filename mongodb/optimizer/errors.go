package optimizer

import (
	"fmt"
)

/*
ErrorType represents the type of error that occurred during optimization
*/
type ErrorType string

const (
	/*
		ErrorTypeConnection represents a connection error
	*/
	ErrorTypeConnection ErrorType = "connection"
	/*
		ErrorTypeIndex represents an index-related error
	*/
	ErrorTypeIndex ErrorType = "index"
	/*
		ErrorTypeQuery represents a query-related error
	*/
	ErrorTypeQuery ErrorType = "query"
	/*
		ErrorTypeSchema represents a schema-related error
	*/
	ErrorTypeSchema ErrorType = "schema"
	/*
		ErrorTypeConfig represents a configuration-related error
	*/
	ErrorTypeConfig ErrorType = "config"
	/*
		ErrorTypeValidation represents a validation error
	*/
	ErrorTypeValidation ErrorType = "validation"
	/*
		ErrorTypeRollback represents a rollback error
	*/
	ErrorTypeRollback ErrorType = "rollback"
	/*
		ErrorTypeUnknown represents an unknown error
	*/
	ErrorTypeUnknown ErrorType = "unknown"
)

/*
OptimizerError represents an error that occurred during optimization
*/
type OptimizerError struct {
	Type        ErrorType
	Message     string
	OriginalErr error
	Database    string
	Collection  string
	Command     string
}

/*
Error implements the error interface
*/
func (e *OptimizerError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s error: %s - %v", e.Type, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

/*
Unwrap returns the original error
*/
func (e *OptimizerError) Unwrap() error {
	return e.OriginalErr
}

// NewOptimizerError creates a new optimizer error
func NewOptimizerError(errType ErrorType, message string, originalErr error) *OptimizerError {
	return &OptimizerError{
		Type:        errType,
		Message:     message,
		OriginalErr: originalErr,
	}
}

// WithDatabase adds database information to the error
func (e *OptimizerError) WithDatabase(db string) *OptimizerError {
	e.Database = db
	return e
}

// WithCollection adds collection information to the error
func (e *OptimizerError) WithCollection(coll string) *OptimizerError {
	e.Collection = coll
	return e
}

// WithCommand adds command information to the error
func (e *OptimizerError) WithCommand(cmd string) *OptimizerError {
	e.Command = cmd
	return e
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	var optimizerErr *OptimizerError
	if err == nil {
		return false
	}

	if e, ok := err.(*OptimizerError); ok {
		optimizerErr = e
	} else {
		return false
	}

	return optimizerErr.Type == ErrorTypeConnection
}

// IsIndexError checks if the error is an index-related error
func IsIndexError(err error) bool {
	var optimizerErr *OptimizerError
	if err == nil {
		return false
	}

	if e, ok := err.(*OptimizerError); ok {
		optimizerErr = e
	} else {
		return false
	}

	return optimizerErr.Type == ErrorTypeIndex
}

// IsRollbackError checks if the error is a rollback error
func IsRollbackError(err error) bool {
	var optimizerErr *OptimizerError
	if err == nil {
		return false
	}

	if e, ok := err.(*OptimizerError); ok {
		optimizerErr = e
	} else {
		return false
	}

	return optimizerErr.Type == ErrorTypeRollback
}
