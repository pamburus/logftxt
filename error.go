package logftxt

import "fmt"

// ---

// ErrFileNotFound is an error that is returned in case file is not found.
//
//nolint:errname // We do not want to break backward compatibility.
type ErrFileNotFound struct {
	Filename string
	cause    error
}

// Error returns error message.
func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("file %q not found", e.Filename)
}

// Is returns true if e is a sub-class of err.
func (e ErrFileNotFound) Is(err error) bool {
	if other, ok := err.(ErrFileNotFound); ok {
		return other.Filename == "" || other.Filename == e.Filename
	}

	return false
}

// Unwrap returns original error.
func (e ErrFileNotFound) Unwrap() error {
	return e.cause
}
