package errors

import "fmt"

type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewValidationError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
