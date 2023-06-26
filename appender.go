// Package logftxt provides logf.Appender and logf.Encoder implementations that output log messages in a textual human-readable form.
package logftxt

import (
	"io"
	"os"

	"github.com/ssgreg/logf"

	"github.com/pamburus/logftxt/internal/pkg/env"
	"github.com/pamburus/logftxt/internal/pkg/tty"
)

// NewAppender returns a new logf.Appender with the given Writer and
// optional custom configuration.
func NewAppender(w io.Writer, options ...AppenderOption) logf.Appender {
	o := defaultAppenderOptions().With(options)
	if o.color == ColorAuto {
		switch env.ColorSetting(o.env) {
		case env.ColorAlways:
			o.color = ColorAlways
		case env.ColorNever:
			o.color = ColorNever
		}
	}

	switch o.color {
	case ColorAlways:
		o.encoderOptions.color = true
	case ColorNever:
		o.encoderOptions.color = false
	default:
		if f, ok := w.(*os.File); ok {
			if tty.EnableSeqTTY(f, true) {
				o.encoderOptions.color = true
			}
		}
	}

	return logf.NewWriteAppender(w, newEncoder(o.encoderOptions))
}
