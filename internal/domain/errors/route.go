package errors

// Route-specific domain errors
var (
	ErrRouteMissingPath = &DomainError{
		Code:    "MISSING_PATH_ERROR",
		Message: "Missing path",
	}

	ErrRouteMissingMethod = &DomainError{
		Code:    "MISSING_METHOD_ERROR",
		Message: "Missing method",
	}

	ErrRouteMissingBackend = &DomainError{
		Code:    "MISSING_BACKEND_ERROR",
		Message: "Missing Backend",
	}
)
