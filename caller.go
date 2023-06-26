package logftxt

import (
	"strconv"

	"github.com/ssgreg/logf"
)

// ---

// CallerEncodeFunc is a function that encodes logf.EntryCaller as a text.
type CallerEncodeFunc func([]byte, logf.EntryCaller) []byte

func (f CallerEncodeFunc) toEncoderOptions(o *encoderOptions) {
	o.encodeCaller = f
}

func (f CallerEncodeFunc) toAppenderOptions(o *appenderOptions) {
	o.encodeCaller = f
}

// ---

// CallerShort returns a CallerEncodeFunc that encodes caller
// keeping only package name, base filename and line number.
func CallerShort() CallerEncodeFunc {
	return func(buf []byte, c logf.EntryCaller) []byte {
		buf = append(buf, c.FileWithPackage()...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(c.Line), 10)

		return buf
	}
}

// CallerLong returns a CallerEncodeFunc that encodes caller keeping full file path and line number.
func CallerLong() CallerEncodeFunc {
	return func(buf []byte, c logf.EntryCaller) []byte {
		buf = append(buf, c.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(c.Line), 10)

		return buf
	}
}

// ---

var _ EncoderOption = CallerEncodeFunc(nil)
var _ AppenderOption = CallerEncodeFunc(nil)
