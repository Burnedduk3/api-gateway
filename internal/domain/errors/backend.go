package errors

// backend-specific domain errors
var (
	ErrBackendMissingHost = &DomainError{
		Code:    "MISSING_HOST_ERROR",
		Message: "Missing host",
	}

	ErrBackendInvalidHost = &DomainError{
		Code:    "INVALID_HOST_ERROR",
		Message: "Invalid host",
	}
)
