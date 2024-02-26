// Package logftxt provides logf.Appender and logf.Encoder implementations that output log messages in a textual human-readable form.
package logftxt

import (
	"io"

	"github.com/ssgreg/logf"

	"github.com/pamburus/ansitty"
)

// NewAppender returns a new logf.Appender with the given Writer and
// optional custom configuration.
func NewAppender(w io.Writer, options ...AppenderOption) logf.Appender {
	o := defaultAppenderOptions().With(options)
	o.color = o.color.resolved(o.env)

	if o.color == ColorAuto {
		if ansitty.Enable(w) {
			o.color = ColorAlways
		}
	}

	return logf.NewWriteAppender(w, newEncoder(o.encoderOptions))
}
