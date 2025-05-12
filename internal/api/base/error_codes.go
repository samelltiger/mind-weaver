package base

// Error codes for API responses
const (
	// Common errors
	ErrCodeSuccess       = 0
	ErrCodeInternalError = 10001
	ErrCodeInvalidParams = 10002
	ErrCodeNotFound      = 10004

	// Connection related errors
	ErrCodeConnectionFail = 20001

	// Task related errors
	ErrCodeTaskNotRunning = 30001
	ErrCodeTaskFailed     = 30002
)

// Error messages mapping
var ErrorMessages = map[int]string{
	ErrCodeSuccess:        "Success",
	ErrCodeInternalError:  "Internal server error",
	ErrCodeInvalidParams:  "Invalid parameters",
	ErrCodeNotFound:       "Resource not found",
	ErrCodeConnectionFail: "Connection failed",
	ErrCodeTaskNotRunning: "Task is not running",
	ErrCodeTaskFailed:     "Task execution failed",
}

// GetErrorMessage returns the error message for a given error code
func GetErrorMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return "Unknown error"
}
