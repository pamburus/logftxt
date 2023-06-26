package logftxt

import (
	"bytes"
	"fmt"
)

// ---

// ErrorShort returns an error formatting function that just returns error message.
func ErrorShort() ErrorEncodeFunc {
	return func(buf []byte, err error) []byte {
		buf = append(buf, err.Error()...)

		return buf
	}
}

// ErrorLong returns an error formatting function that uses "%+v" format and can be used
// for errors with additional information like those provided by github.com/pkg/errors package.
func ErrorLong() ErrorEncodeFunc {
	return func(buf []byte, err error) []byte {
		w := bytes.NewBuffer(buf)
		_, _ = fmt.Fprintf(w, "%+v", err)

		return w.Bytes()
	}
}

// ---

// ErrorEncodeFunc is a function that encodes errors as a text.
type ErrorEncodeFunc func([]byte, error) []byte

func (f ErrorEncodeFunc) toEncoderOptions(o *encoderOptions) {
	o.encodeError = f
}

func (f ErrorEncodeFunc) toAppenderOptions(o *appenderOptions) {
	o.encodeError = f
}
