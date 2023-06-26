package logftxt

import "time"

// ---

// TimeLayout returns TimeEncodeFunc that encodes time using the given layout.
func TimeLayout(layout string) TimeEncodeFunc {
	return func(buf []byte, t time.Time) []byte {
		return t.AppendFormat(buf, layout)
	}
}

// ---

// TimeEncodeFunc is a function that encodes time.Time into text.
type TimeEncodeFunc func([]byte, time.Time) []byte

// Timestamp returns a TimestampEncodeFunc that encodes timestamp into text the same way.
func (f TimeEncodeFunc) Timestamp() TimestampEncodeFunc {
	return TimestampEncodeFunc(f)
}

// TimeValue returns a TimeValueEncodeFunc that encodes time values into text the same way.
func (f TimeEncodeFunc) TimeValue() TimeValueEncodeFunc {
	return TimeValueEncodeFunc(f)
}

// ---

// TimestampEncodeFunc is a function that encodes log message timestamp into text.
type TimestampEncodeFunc TimeEncodeFunc

func (f TimestampEncodeFunc) toEncoderOptions(o *encoderOptions) {
	o.encodeTimestamp = f
}

func (f TimestampEncodeFunc) toAppenderOptions(o *appenderOptions) {
	o.encodeTimestamp = f
}

// ---

// TimeValueEncodeFunc is a function that encodes time field values into text.
type TimeValueEncodeFunc TimeEncodeFunc

func (f TimeValueEncodeFunc) toEncoderOptions(o *encoderOptions) {
	o.encodeTimeValue = f
}

func (f TimeValueEncodeFunc) toAppenderOptions(o *appenderOptions) {
	o.encodeTimeValue = f
}

// ---

var (
	_ AppenderOption = TimestampEncodeFunc(nil)
	_ EncoderOption  = TimestampEncodeFunc(nil)
	_ AppenderOption = TimeValueEncodeFunc(nil)
	_ EncoderOption  = TimeValueEncodeFunc(nil)
)
