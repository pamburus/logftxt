package logftxt

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/ssgreg/logf"
)

// NewEncoder constructs a new logf.Encoder that encodes log messages in a human-readable text representation.
func NewEncoder(options ...EncoderOption) logf.Encoder {
	return newEncoder(defaultEncoderOptions().With(options))
}

// ---

type encoder struct {
	encoderOptions
	cfg   *Config
	theme *Theme
	pool  chan *entryEncoder
	once  sync.Once
}

func (e *encoder) Encode(buf *logf.Buffer, entry logf.Entry) error {
	e.once.Do(func() {
		e.setup(buf, entry.Time)
	})

	ee := e.getEntryEncoder()
	ee.buf = buf
	ee.entry = entry
	ee.startBufLen = buf.Len()

	err := ee.encode()

	e.putEntryEncoder(ee)

	return err
}

func (e *encoder) getEntryEncoder() *entryEncoder {
	select {
	case ee := <-e.pool:
		return ee
	default:
		return &entryEncoder{
			e.encoderOptions,
			e.theme,
			logf.Entry{},
			nil,
			logf.NewCache(100),
			0,
			nil,
			0,
			0,
			newStyler().Disabled(e.color == ColorNever),
		}
	}
}

func (e *encoder) putEntryEncoder(ee *entryEncoder) {
	ee.entry = logf.Entry{}
	ee.buf = nil

	select {
	case e.pool <- ee:
	default:
	}
}

func (e *encoder) setup(buf *logf.Buffer, ts time.Time) {
	var messages []logf.Entry

	setupContext := domain{e.env, e.fs}

	for i := len(e.provideConfig) - 1; i >= 0; i-- {
		cfg, err := e.provideConfig[i](setupContext)
		if err != nil {
			messages = append(messages, logf.Entry{
				Text:   "failed to load configuration file so using previous defaults",
				Fields: []logf.Field{logf.Error(err)},
			})
		} else if cfg != nil {
			e.cfg = cfg

			break
		}
	}
	if e.cfg == nil {
		e.cfg = DefaultConfig()
	}

	if e.cfg.Theme.name != "" {
		e.provideTheme = append([]ThemeProvideFunc{e.cfg.Theme.fn()}, e.provideTheme...)
	}
	for i := len(e.provideTheme) - 1; i >= 0; i-- {
		theme, err := e.provideTheme[i](setupContext)
		if err != nil {
			messages = append(messages, logf.Entry{
				Text:   "failed to setup preferred theme so using previous defaults",
				Fields: []logf.Field{logf.Error(err)},
			})
		} else if theme != nil {
			e.theme = theme

			break
		}
	}
	if e.theme == nil {
		e.theme = DefaultTheme()
	}

	if e.encodeDuration == nil {
		switch e.cfg.Values.Duration.Format {
		case DurationFormatSeconds:
			e.encodeDuration = DurationAsSeconds(e.cfg.Values.Duration.Precision)
		case DurationFormatDynamic:
			e.encodeDuration = DurationAsText()
		case DurationFormatDefault, DurationFormatHMS:
			fallthrough
		default:
			e.encodeDuration = DurationAsHMS(e.cfg.Values.Duration.Precision)
		}
	}

	if e.encodeTimestamp == nil {
		e.encodeTimestamp = TimestampEncodeFunc(TimeLayout(e.cfg.Timestamp.Format))
	}

	if e.encodeTimeValue == nil {
		e.encodeTimeValue = TimeValueEncodeFunc(TimeLayout(e.cfg.Timestamp.Format))
	}

	if e.encodeCaller == nil {
		switch e.cfg.Caller.Format {
		case CallerFormatLong:
			e.encodeCaller = CallerLong()
		case CallerFormatShort, CallerFormatDefault:
			fallthrough
		default:
			e.encodeCaller = CallerShort()
		}
	}

	if e.encodeError == nil {
		switch e.cfg.Values.Error.Format {
		case ErrorFormatLong:
			e.encodeError = ErrorLong()
		case ErrorFormatShort, ErrorFormatDefault:
			fallthrough
		default:
			e.encodeError = ErrorShort()
		}
	}

	for _, message := range messages {
		message.Time = ts.Add(-time.Nanosecond)
		message.Level = logf.LevelWarn

		e.log(buf, message)
	}
}

func (e *encoder) log(buf *logf.Buffer, entry logf.Entry) {
	ee := e.getEntryEncoder()
	ee.buf = buf
	ee.startBufLen = buf.Len()
	ee.entry = entry
	ee.entry.LoggerName = loggerName
	ee.entry.LoggerID = rand.Int31() //nolint:gosec

	_ = ee.encode()

	e.putEntryEncoder(ee)
}

// ---

type entryEncoder struct {
	encoderOptions
	theme       *Theme
	entry       logf.Entry
	buf         *logf.Buffer
	cache       *logf.Cache
	startBufLen int
	objectKeys  []string
	objectScope int
	lastPos     int

	styler styler
}

func (e *entryEncoder) encode() error {
	for _, item := range e.theme.items {
		pos := e.appendSeparator()
		item.encode(e)
		e.confirmSeparator(pos)
	}

	e.buf.AppendByte('\n')

	return nil
}

func (e *entryEncoder) EncodeFieldAny(k string, v interface{}) {
	e.appendField(k, func() {
		e.EncodeTypeAny(v)
	})
}

func (e *entryEncoder) EncodeFieldBool(k string, v bool) {
	e.appendField(k, func() {
		e.EncodeTypeBool(v)
	})
}

func (e *entryEncoder) EncodeFieldInt64(k string, v int64) {
	e.appendField(k, func() {
		e.EncodeTypeInt64(v)
	})
}

func (e *entryEncoder) EncodeFieldInt32(k string, v int32) {
	e.appendField(k, func() {
		e.EncodeTypeInt32(v)
	})
}

func (e *entryEncoder) EncodeFieldInt16(k string, v int16) {
	e.appendField(k, func() {
		e.EncodeTypeInt16(v)
	})
}

func (e *entryEncoder) EncodeFieldInt8(k string, v int8) {
	e.appendField(k, func() {
		e.EncodeTypeInt8(v)
	})
}

func (e *entryEncoder) EncodeFieldUint64(k string, v uint64) {
	e.appendField(k, func() {
		e.EncodeTypeUint64(v)
	})
}

func (e *entryEncoder) EncodeFieldUint32(k string, v uint32) {
	e.appendField(k, func() {
		e.EncodeTypeUint32(v)
	})
}

func (e *entryEncoder) EncodeFieldUint16(k string, v uint16) {
	e.appendField(k, func() {
		e.EncodeTypeUint16(v)
	})
}

func (e *entryEncoder) EncodeFieldUint8(k string, v uint8) {
	e.appendField(k, func() {
		e.EncodeTypeUint8(v)
	})
}

func (e *entryEncoder) EncodeFieldFloat64(k string, v float64) {
	e.appendField(k, func() {
		e.EncodeTypeFloat64(v)
	})
}

func (e *entryEncoder) EncodeFieldFloat32(k string, v float32) {
	e.appendField(k, func() {
		e.EncodeTypeFloat32(v)
	})
}

func (e *entryEncoder) EncodeFieldString(k string, v string) {
	e.appendField(k, func() {
		e.EncodeTypeString(v)
	})
}

func (e *entryEncoder) EncodeFieldDuration(k string, v time.Duration) {
	e.appendField(k, func() {
		e.EncodeTypeDuration(v)
	})
}

func (e *entryEncoder) EncodeFieldError(k string, v error) {
	e.appendField(k, func() {
		e.EncodeTypeError(v)
	})
}

func (e *entryEncoder) EncodeFieldTime(k string, v time.Time) {
	e.appendField(k, func() {
		e.EncodeTypeTime(v)
	})
}

func (e *entryEncoder) EncodeFieldArray(k string, v logf.ArrayEncoder) {
	e.appendField(k, func() {
		e.EncodeTypeArray(v)
	})
}

func (e *entryEncoder) EncodeFieldObject(k string, v logf.ObjectEncoder) {
	if e.flattenObjects {
		e.objectKeys = append(e.objectKeys, k)
		oe := objectEncoder{e, 0}
		_ = v.EncodeLogfObject(&oe)
		e.objectKeys = e.objectKeys[:len(e.objectKeys)-1]
	} else {
		e.appendField(k, func() {
			e.EncodeTypeObject(v)
		})
	}
}

func (e *entryEncoder) EncodeFieldBytes(k string, v []byte) {
	e.appendField(k, func() {
		e.EncodeTypeBytes(v)
	})
}

func (e *entryEncoder) EncodeFieldBools(k string, v []bool) {
	e.appendField(k, func() {
		e.EncodeTypeBools(v)
	})
}

func (e *entryEncoder) EncodeFieldStrings(k string, v []string) {
	e.appendField(k, func() {
		e.EncodeTypeStrings(v)
	})
}

func (e *entryEncoder) EncodeFieldInts64(k string, v []int64) {
	e.appendField(k, func() {
		e.EncodeTypeInts64(v)
	})
}

func (e *entryEncoder) EncodeFieldInts32(k string, v []int32) {
	e.appendField(k, func() {
		e.EncodeTypeInts32(v)
	})
}

func (e *entryEncoder) EncodeFieldInts16(k string, v []int16) {
	e.appendField(k, func() {
		e.EncodeTypeInts16(v)
	})
}

func (e *entryEncoder) EncodeFieldInts8(k string, v []int8) {
	e.appendField(k, func() {
		e.EncodeTypeInts8(v)
	})
}

func (e *entryEncoder) EncodeFieldUints64(k string, v []uint64) {
	e.appendField(k, func() {
		e.EncodeTypeUints64(v)
	})
}

func (e *entryEncoder) EncodeFieldUints32(k string, v []uint32) {
	e.appendField(k, func() {
		e.EncodeTypeUints32(v)
	})
}

func (e *entryEncoder) EncodeFieldUints16(k string, v []uint16) {
	e.appendField(k, func() {
		e.EncodeTypeUints16(v)
	})
}

func (e *entryEncoder) EncodeFieldUints8(k string, v []uint8) {
	e.appendField(k, func() {
		e.EncodeTypeUints8(v)
	})
}

func (e *entryEncoder) EncodeFieldFloats64(k string, v []float64) {
	e.appendField(k, func() {
		e.EncodeTypeFloats64(v)
	})
}

func (e *entryEncoder) EncodeFieldFloats32(k string, v []float32) {
	e.appendField(k, func() {
		e.EncodeTypeFloats32(v)
	})
}

func (e *entryEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	e.appendField(k, func() {
		e.EncodeTypeDurations(v)
	})
}

//nolint:gocyclo
func (e *entryEncoder) EncodeTypeAny(v interface{}) {
	switch v := v.(type) {
	case nil:
		e.theme.fmt.Null.encode(e, func() {
			e.buf.AppendString("null")
		})
	case bool:
		e.EncodeTypeBool(v)
	case int:
		e.EncodeTypeInt64(int64(v))
	case int64:
		e.EncodeTypeInt64(v)
	case int32:
		e.EncodeTypeInt32(v)
	case int16:
		e.EncodeTypeInt16(v)
	case int8:
		e.EncodeTypeInt8(v)
	case uint:
		e.EncodeTypeUint64(uint64(v))
	case uint64:
		e.EncodeTypeUint64(v)
	case uint32:
		e.EncodeTypeUint32(v)
	case uint16:
		e.EncodeTypeUint16(v)
	case uint8:
		e.EncodeTypeUint8(v)
	case float64:
		e.EncodeTypeFloat64(v)
	case float32:
		e.EncodeTypeFloat32(v)
	case time.Time:
		e.EncodeTypeTime(v)
	case time.Duration:
		e.EncodeTypeDuration(v)
	case logf.ArrayEncoder:
		e.EncodeTypeArray(v)
	case logf.ObjectEncoder:
		e.EncodeTypeObject(v)
	case []byte:
		e.EncodeTypeBytes(v)
	case []string:
		e.EncodeTypeStrings(v)
	case []bool:
		e.EncodeTypeBools(v)
	case []int:
		e.appendArray(len(v), func(i int) {
			e.EncodeTypeInt64(int64(v[i]))
		})
	case []int64:
		e.EncodeTypeInts64(v)
	case []int32:
		e.EncodeTypeInts32(v)
	case []int16:
		e.EncodeTypeInts16(v)
	case []int8:
		e.EncodeTypeInts8(v)
	case []uint:
		e.appendArray(len(v), func(i int) {
			e.EncodeTypeUint64(uint64(v[i]))
		})
	case []uint64:
		e.EncodeTypeUints64(v)
	case []uint32:
		e.EncodeTypeUints32(v)
	case []uint16:
		e.EncodeTypeUints16(v)
	case []float64:
		e.EncodeTypeFloats64(v)
	case []float32:
		e.EncodeTypeFloats32(v)
	case []time.Duration:
		e.EncodeTypeDurations(v)
	case string:
		e.EncodeTypeString(v)
	case fmt.Stringer:
		e.EncodeTypeString(v.String())
	case error:
		e.EncodeTypeError(v)

	default:
		rv := reflect.ValueOf(v)
		switch rv.Type().Kind() {
		case reflect.String:
			e.EncodeTypeString(rv.String())
		case reflect.Bool:
			e.EncodeTypeBool(rv.Bool())
		case reflect.Int:
			e.EncodeTypeInt64(rv.Int())
		case reflect.Int8:
			e.EncodeTypeInt8(int8(rv.Int()))
		case reflect.Int16:
			e.EncodeTypeInt16(int16(rv.Int()))
		case reflect.Int32:
			e.EncodeTypeInt32(int32(rv.Int()))
		case reflect.Int64:
			e.EncodeTypeInt64(rv.Int())
		case reflect.Uint:
			e.EncodeTypeUint64(rv.Uint())
		case reflect.Uint8:
			e.EncodeTypeUint8(uint8(rv.Uint()))
		case reflect.Uint16:
			e.EncodeTypeUint16(uint16(rv.Uint()))
		case reflect.Uint32:
			e.EncodeTypeUint32(uint32(rv.Uint()))
		case reflect.Uint64:
			e.EncodeTypeUint64(rv.Uint())
		case reflect.Float32:
			e.EncodeTypeFloat32(float32(rv.Float()))
		case reflect.Float64:
			e.EncodeTypeFloat64(rv.Float())
		case reflect.Slice, reflect.Array:
			e.EncodeTypeArray(anyArray{rv})
		default:
			_, _ = fmt.Fprintf(e.buf, "%v", v)
		}
	}
}

func (e *entryEncoder) EncodeTypeBool(v bool) {
	e.theme.fmt.Boolean.encode(e, func() {
		logf.AppendBool(e.buf, v)
	})
}

func (e *entryEncoder) EncodeTypeInt64(v int64) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendInt(e.buf, v)
	})
}

func (e *entryEncoder) EncodeTypeInt32(v int32) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendInt(e.buf, int64(v))
	})
}

func (e *entryEncoder) EncodeTypeInt16(v int16) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendInt(e.buf, int64(v))
	})
}

func (e *entryEncoder) EncodeTypeInt8(v int8) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendInt(e.buf, int64(v))
	})
}

func (e *entryEncoder) EncodeTypeUint64(v uint64) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendUint(e.buf, v)
	})
}

func (e *entryEncoder) EncodeTypeUint32(v uint32) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendUint(e.buf, uint64(v))
	})
}

func (e *entryEncoder) EncodeTypeUint16(v uint16) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendUint(e.buf, uint64(v))
	})
}

func (e *entryEncoder) EncodeTypeUint8(v uint8) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendUint(e.buf, uint64(v))
	})
}

func (e *entryEncoder) EncodeTypeFloat64(v float64) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendFloat64(e.buf, v)
	})
}

func (e *entryEncoder) EncodeTypeFloat32(v float32) {
	e.theme.fmt.Number.encode(e, func() {
		logf.AppendFloat32(e.buf, v)
	})
}

func (e *entryEncoder) EncodeTypeDuration(v time.Duration) {
	e.theme.fmt.Duration.encode(e, func() {
		e.buf.Data = e.encodeDuration(e.buf.Data, v)
	})
}

func (e *entryEncoder) EncodeTypeTime(v time.Time) {
	e.theme.fmt.Time.encode(e, func() {
		e.appendTimeValue(v)
	})
}

func (e *entryEncoder) EncodeTypeString(v string) {
	e.theme.fmt.String.encode(e, func() {
		if v == "null" {
			e.buf.AppendString(`"null"`)
		} else {
			e.appendAutoQuotedString(v)
		}
	})
}

func (e *entryEncoder) EncodeTypeStrings(v []string) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeString(v[i])
	})
}

func (e *entryEncoder) EncodeTypeBytes(v []byte) {
	e.theme.fmt.String.encode(e, func() {
		encoding := base64.URLEncoding
		encoding.Encode(e.buf.ExtendBytes(encoding.EncodedLen(len(v))), v)
	})
}

func (e *entryEncoder) EncodeTypeBools(v []bool) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeBool(v[i])
	})
}

func (e *entryEncoder) EncodeTypeInts64(v []int64) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeInt64(v[i])
	})
}

func (e *entryEncoder) EncodeTypeInts32(v []int32) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeInt32(v[i])
	})
}

func (e *entryEncoder) EncodeTypeInts16(v []int16) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeInt16(v[i])
	})
}

func (e *entryEncoder) EncodeTypeInts8(v []int8) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeInt8(v[i])
	})
}

func (e *entryEncoder) EncodeTypeUints64(v []uint64) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeUint64(v[i])
	})
}

func (e *entryEncoder) EncodeTypeUints32(v []uint32) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeUint32(v[i])
	})
}

func (e *entryEncoder) EncodeTypeUints16(v []uint16) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeUint16(v[i])
	})
}

func (e *entryEncoder) EncodeTypeUints8(v []uint8) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeUint8(v[i])
	})
}

func (e *entryEncoder) EncodeTypeFloats64(v []float64) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeFloat64(v[i])
	})
}

func (e *entryEncoder) EncodeTypeFloats32(v []float32) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeFloat32(v[i])
	})
}

func (e *entryEncoder) EncodeTypeDurations(v []time.Duration) {
	e.appendArray(len(v), func(i int) {
		e.EncodeTypeDuration(v[i])
	})
}

func (e *entryEncoder) EncodeTypeArray(v logf.ArrayEncoder) {
	old := e.objectScope
	e.objectScope = len(e.objectKeys)

	e.theme.fmt.Array.encode(e, func() {
		e.buf.AppendString(e.theme.fmt.Array.inner.prefix)
		ae := arrayEncoder{e, 0}
		_ = v.EncodeLogfArray(&ae)
		if ae.n == 0 {
			e.buf.Data = e.buf.Data[:e.buf.Len()-len(e.theme.fmt.Array.inner.prefix)]
		} else {
			e.buf.AppendString(e.theme.fmt.Array.inner.suffix)
		}
	})

	e.objectScope = old
}

func (e *entryEncoder) EncodeTypeObject(v logf.ObjectEncoder) {
	e.theme.fmt.Object.encode(e, func() {
		e.buf.AppendString(e.theme.fmt.Object.inner.prefix)
		oe := objectEncoder{e, 0}
		_ = v.EncodeLogfObject(&oe)
		if oe.n == 0 {
			e.buf.Data = e.buf.Data[:e.buf.Len()-len(e.theme.fmt.Object.inner.prefix)]
		} else {
			e.buf.AppendString(e.theme.fmt.Object.inner.suffix)
		}
	})
}

func (e *entryEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	e.buf.Data = append(e.buf.Data, *(*[]byte)(v)...)
}

func (e *entryEncoder) EncodeTypeError(v error) {
	e.theme.fmt.Error.encode(e, func() {
		e.buf.AppendString(e.theme.fmt.Error.inner.prefix)
		e.appendError(v)
		e.buf.AppendString(e.theme.fmt.Error.inner.suffix)
	})
}

func (e *entryEncoder) appendArray(n int, appendElement func(int)) {
	e.theme.fmt.Array.encode(e, func() {
		if n != 0 {
			e.buf.AppendString(e.theme.fmt.Array.inner.prefix)
			for i := 0; i != n; i++ {
				if i != 0 {
					e.theme.fmt.Array.separator.encode(e)
				}
				appendElement(i)
			}
			e.buf.AppendString(e.theme.fmt.Array.inner.suffix)
		}
	})
}

func (e *entryEncoder) appendField(k string, appendValue func()) {
	e.styler.Use(e.theme.fmt.Field.inner.style, e.buf, func() {
		e.addKey(k)
		e.theme.fmt.Field.separator.encode(e)
		appendValue()
	})
}

func (e *entryEncoder) appendSeparator() int {
	if !e.empty() && e.buf.Len() == e.lastPos {
		e.buf.AppendByte(' ')
	}

	return e.buf.Len()
}

func (e *entryEncoder) appendCustomSeparator(fn func()) int {
	if !e.empty() && e.buf.Len() == e.lastPos {
		fn()
	}

	return e.buf.Len()
}

func (e *entryEncoder) confirmSeparator(start int) bool {
	if e.buf.Len() == start && !e.empty() {
		e.buf.Data = e.buf.Data[:e.lastPos]

		return false
	}

	e.lastPos = e.buf.Len()

	return true
}

func (e *entryEncoder) empty() bool {
	return e.buf.Len() == e.startBufLen
}

func (e *entryEncoder) addKey(k string) {
	e.theme.fmt.Key.encode(e, func() {
		for _, prefix := range e.objectKeys[e.objectScope:] {
			e.appendAutoQuotedString(prefix)
			e.theme.fmt.Key.separator.encode(e)
		}
		e.appendAutoQuotedString(k)
	})
}

func (e *entryEncoder) appendAutoQuotedString(v string) {
	switch {
	case len(v) == 0:
		e.buf.AppendString(`""`)
	case stringNeedsQuoting(v):
		e.buf.AppendByte('"')
		_ = logf.EscapeString(e.buf, v)
		e.buf.AppendByte('"')
	default:
		e.buf.AppendString(v)
	}
}

func (e *entryEncoder) appendTimestamp(t time.Time) {
	e.buf.Data = e.encodeTimestamp(e.buf.Data, t)
}

func (e *entryEncoder) appendTimeValue(t time.Time) {
	e.buf.Data = e.encodeTimeValue(e.buf.Data, t)
}

func (e *entryEncoder) appendCaller(caller logf.EntryCaller) {
	e.buf.Data = e.encodeCaller(e.buf.Data, caller)
}

func (e *entryEncoder) appendError(v error) {
	e.buf.Data = e.encodeError(e.buf.Data, v)
}

// ---

type arrayEncoder struct {
	e *entryEncoder
	n int
}

func (e *arrayEncoder) EncodeTypeAny(v interface{}) {
	e.encode(func() {
		e.e.EncodeTypeAny(v)
	})
}

func (e *arrayEncoder) EncodeTypeBool(v bool) {
	e.encode(func() {
		e.e.EncodeTypeBool(v)
	})
}

func (e *arrayEncoder) EncodeTypeInt64(v int64) {
	e.encode(func() {
		e.e.EncodeTypeInt64(v)
	})
}

func (e *arrayEncoder) EncodeTypeInt32(v int32) {
	e.encode(func() {
		e.e.EncodeTypeInt32(v)
	})
}

func (e *arrayEncoder) EncodeTypeInt16(v int16) {
	e.encode(func() {
		e.e.EncodeTypeInt16(v)
	})
}

func (e *arrayEncoder) EncodeTypeInt8(v int8) {
	e.encode(func() {
		e.e.EncodeTypeInt8(v)
	})
}

func (e *arrayEncoder) EncodeTypeUint64(v uint64) {
	e.encode(func() {
		e.e.EncodeTypeUint64(v)
	})
}

func (e *arrayEncoder) EncodeTypeUint32(v uint32) {
	e.encode(func() {
		e.e.EncodeTypeUint32(v)
	})
}

func (e *arrayEncoder) EncodeTypeUint16(v uint16) {
	e.encode(func() {
		e.e.EncodeTypeUint16(v)
	})
}

func (e *arrayEncoder) EncodeTypeUint8(v uint8) {
	e.encode(func() {
		e.e.EncodeTypeUint8(v)
	})
}

func (e *arrayEncoder) EncodeTypeFloat64(v float64) {
	e.encode(func() {
		e.e.EncodeTypeFloat64(v)
	})
}

func (e *arrayEncoder) EncodeTypeFloat32(v float32) {
	e.encode(func() {
		e.e.EncodeTypeFloat32(v)
	})
}

func (e *arrayEncoder) EncodeTypeDuration(v time.Duration) {
	e.encode(func() {
		e.e.EncodeTypeDuration(v)
	})
}

func (e *arrayEncoder) EncodeTypeTime(v time.Time) {
	e.encode(func() {
		e.e.EncodeTypeTime(v)
	})
}

func (e *arrayEncoder) EncodeTypeString(v string) {
	e.encode(func() {
		e.e.EncodeTypeString(v)
	})
}

func (e *arrayEncoder) EncodeTypeStrings(v []string) {
	e.encode(func() {
		e.e.EncodeTypeStrings(v)
	})
}

func (e *arrayEncoder) EncodeTypeBytes(v []byte) {
	e.encode(func() {
		e.e.EncodeTypeBytes(v)
	})
}

func (e *arrayEncoder) EncodeTypeBools(v []bool) {
	e.encode(func() {
		e.e.EncodeTypeBools(v)
	})
}

func (e *arrayEncoder) EncodeTypeInts64(v []int64) {
	e.encode(func() {
		e.e.EncodeTypeInts64(v)
	})
}

func (e *arrayEncoder) EncodeTypeInts32(v []int32) {
	e.encode(func() {
		e.e.EncodeTypeInts32(v)
	})
}

func (e *arrayEncoder) EncodeTypeInts16(v []int16) {
	e.encode(func() {
		e.e.EncodeTypeInts16(v)
	})
}

func (e *arrayEncoder) EncodeTypeInts8(v []int8) {
	e.encode(func() {
		e.e.EncodeTypeInts8(v)
	})
}

func (e *arrayEncoder) EncodeTypeUints64(v []uint64) {
	e.encode(func() {
		e.e.EncodeTypeUints64(v)
	})
}

func (e *arrayEncoder) EncodeTypeUints32(v []uint32) {
	e.encode(func() {
		e.e.EncodeTypeUints32(v)
	})
}

func (e *arrayEncoder) EncodeTypeUints16(v []uint16) {
	e.encode(func() {
		e.e.EncodeTypeUints16(v)
	})
}

func (e *arrayEncoder) EncodeTypeUints8(v []uint8) {
	e.encode(func() {
		e.e.EncodeTypeUints8(v)
	})
}

func (e *arrayEncoder) EncodeTypeFloats64(v []float64) {
	e.encode(func() {
		e.e.EncodeTypeFloats64(v)
	})
}

func (e *arrayEncoder) EncodeTypeFloats32(v []float32) {
	e.encode(func() {
		e.e.EncodeTypeFloats32(v)
	})
}

func (e *arrayEncoder) EncodeTypeDurations(v []time.Duration) {
	e.encode(func() {
		e.e.EncodeTypeDurations(v)
	})
}

func (e *arrayEncoder) EncodeTypeArray(v logf.ArrayEncoder) {
	e.encode(func() {
		e.e.EncodeTypeArray(v)
	})
}

func (e *arrayEncoder) EncodeTypeObject(v logf.ObjectEncoder) {
	e.encode(func() {
		e.e.EncodeTypeObject(v)
	})
}

func (e *arrayEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	e.encode(func() {
		e.e.EncodeTypeUnsafeBytes(v)
	})
}

func (e *arrayEncoder) encode(encodeValue func()) {
	if e.n != 0 {
		e.e.theme.fmt.Array.separator.encode(e.e)
	}
	encodeValue()
	e.n++
}

// ---

type objectEncoder struct {
	e *entryEncoder
	n int
}

func (e *objectEncoder) EncodeFieldAny(k string, v interface{}) {
	e.encode(func() {
		e.e.EncodeFieldAny(k, v)
	})
}

func (e *objectEncoder) EncodeFieldBool(k string, v bool) {
	e.encode(func() {
		e.e.EncodeFieldBool(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInt64(k string, v int64) {
	e.encode(func() {
		e.e.EncodeFieldInt64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInt32(k string, v int32) {
	e.encode(func() {
		e.e.EncodeFieldInt32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInt16(k string, v int16) {
	e.encode(func() {
		e.e.EncodeFieldInt16(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInt8(k string, v int8) {
	e.encode(func() {
		e.e.EncodeFieldInt8(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUint64(k string, v uint64) {
	e.encode(func() {
		e.e.EncodeFieldUint64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUint32(k string, v uint32) {
	e.encode(func() {
		e.e.EncodeFieldUint32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUint16(k string, v uint16) {
	e.encode(func() {
		e.e.EncodeFieldUint16(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUint8(k string, v uint8) {
	e.encode(func() {
		e.e.EncodeFieldUint8(k, v)
	})
}

func (e *objectEncoder) EncodeFieldFloat64(k string, v float64) {
	e.encode(func() {
		e.e.EncodeFieldFloat64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldFloat32(k string, v float32) {
	e.encode(func() {
		e.e.EncodeFieldFloat32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldDuration(k string, v time.Duration) {
	e.encode(func() {
		e.e.EncodeFieldDuration(k, v)
	})
}

func (e *objectEncoder) EncodeFieldError(k string, v error) {
	e.encode(func() {
		e.e.EncodeFieldError(k, v)
	})
}

func (e *objectEncoder) EncodeFieldTime(k string, v time.Time) {
	e.encode(func() {
		e.e.EncodeFieldTime(k, v)
	})
}

func (e *objectEncoder) EncodeFieldString(k string, v string) {
	e.encode(func() {
		e.e.EncodeFieldString(k, v)
	})
}

func (e *objectEncoder) EncodeFieldStrings(k string, v []string) {
	e.encode(func() {
		e.e.EncodeFieldStrings(k, v)
	})
}

func (e *objectEncoder) EncodeFieldBytes(k string, v []byte) {
	e.encode(func() {
		e.e.EncodeFieldBytes(k, v)
	})
}

func (e *objectEncoder) EncodeFieldBools(k string, v []bool) {
	e.encode(func() {
		e.e.EncodeFieldBools(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInts64(k string, v []int64) {
	e.encode(func() {
		e.e.EncodeFieldInts64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInts32(k string, v []int32) {
	e.encode(func() {
		e.e.EncodeFieldInts32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInts16(k string, v []int16) {
	e.encode(func() {
		e.e.EncodeFieldInts16(k, v)
	})
}

func (e *objectEncoder) EncodeFieldInts8(k string, v []int8) {
	e.encode(func() {
		e.e.EncodeFieldInts8(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUints64(k string, v []uint64) {
	e.encode(func() {
		e.e.EncodeFieldUints64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUints32(k string, v []uint32) {
	e.encode(func() {
		e.e.EncodeFieldUints32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUints16(k string, v []uint16) {
	e.encode(func() {
		e.e.EncodeFieldUints16(k, v)
	})
}

func (e *objectEncoder) EncodeFieldUints8(k string, v []uint8) {
	e.encode(func() {
		e.e.EncodeFieldUints8(k, v)
	})
}

func (e *objectEncoder) EncodeFieldFloats64(k string, v []float64) {
	e.encode(func() {
		e.e.EncodeFieldFloats64(k, v)
	})
}

func (e *objectEncoder) EncodeFieldFloats32(k string, v []float32) {
	e.encode(func() {
		e.e.EncodeFieldFloats32(k, v)
	})
}

func (e *objectEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	e.encode(func() {
		e.e.EncodeFieldDurations(k, v)
	})
}

func (e *objectEncoder) EncodeFieldArray(k string, v logf.ArrayEncoder) {
	e.encode(func() {
		e.e.EncodeFieldArray(k, v)
	})
}

func (e *objectEncoder) EncodeFieldObject(k string, v logf.ObjectEncoder) {
	e.encode(func() {
		e.e.EncodeFieldObject(k, v)
	})
}

func (e *objectEncoder) encode(encodeField func()) {
	var pos int
	if e.e.flattenObjects && e.e.objectScope == 0 {
		pos = e.e.appendSeparator()
	} else {
		pos = e.e.appendCustomSeparator(func() {
			e.e.theme.fmt.Object.separator.encode(e.e)
		})
	}
	encodeField()
	if e.e.confirmSeparator(pos) {
		e.n++
	}
}

// ---

type anyArray struct {
	v reflect.Value
}

func (a anyArray) EncodeLogfArray(enc logf.TypeEncoder) error {
	for i := 0; i != a.v.Len(); i++ {
		enc.EncodeTypeAny(a.v.Index(i).Interface())
	}

	return nil
}

// ---

func stringNeedsQuoting(s string) bool {
	looksLikeNumber := true
	nDots := 0

	for _, r := range s {
		switch r {
		case '.':
			nDots++
		case '=', '"', ' ', utf8.RuneError:
			return true
		default:
			if r < ' ' {
				return true
			}
			if !isDigit(r) {
				looksLikeNumber = false
			}
		}
	}

	return looksLikeNumber && nDots <= 1
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// ---

func newEncoder(options encoderOptions) *encoder {
	options.color = options.color.resolved(options.env)

	return &encoder{
		options,
		nil,
		nil,
		make(chan *entryEncoder, options.poolSizeLimit),
		sync.Once{},
	}
}

// ---

const loggerName = "logftxt"
