package errors

import (
	"log"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Error ...
type Error struct {
	Type          string
	originalError error
	error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ErrorTraceLogField ...
func ErrorTraceLogField(err *Error) zap.Field {
	return zap.String("error_stacktrace", getTrace(err))
}

// getTrace ...
func getTrace(err *Error) string {
	var stackTrace string
	if err, ok := err.originalError.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			s, err := f.MarshalText()
			if err != nil {
				log.Panic(err.Error())
			}

			stackTrace = stackTrace + string(s) + "\n"
		}
	}

	return stackTrace
}

// Wrap ...
func Wrap(err error, errorType string) *Error {
	return &Error{
		Type:          errorType,
		originalError: errors.Wrap(err, ""),
	}
}

// New ...
func New(message string, errorType string) *Error {
	return &Error{
		Type:          errorType,
		originalError: errors.New(message),
	}
}

func (e *Error) Error() string {
	return e.originalError.Error()
}
