package logftxt

import (
	"github.com/pamburus/logftxt/internal/pkg/env"
)

// ---

// Valid values for ColorSetting.
const (
	ColorAuto ColorSetting = iota
	ColorNever
	ColorAlways
)

// ColorSetting allows to explicitly specify preference on using colors and other ANSI SGR escape sequences in output.
type ColorSetting int

func (s ColorSetting) toAppenderOptions(o *appenderOptions) {
	o.color = s
}

func (s ColorSetting) toEncoderOptions(o *encoderOptions) {
	o.color = s
}

func (s ColorSetting) resolved(e Environment) ColorSetting {
	if s == ColorAuto {
		switch env.ColorSetting(e) {
		case env.ColorAlways:
			s = ColorAlways
		case env.ColorNever:
			s = ColorNever
		}
	}

	return s
}

// ---

// FlattenObjects tells encoder wether to flatten nested objects when encoding.
func FlattenObjects(flatten bool) FlattenObjectsSetting {
	return FlattenObjectsSetting(flatten)
}

// FlattenObjectsSetting tells encoder to flatten or not nested objects when encoding.
type FlattenObjectsSetting bool

func (s FlattenObjectsSetting) toEncoderOptions(o *encoderOptions) {
	o.flattenObjects = bool(s)
}

func (s FlattenObjectsSetting) toAppenderOptions(o *appenderOptions) {
	o.flattenObjects = bool(s)
}

// ---

// AppenderOption is an optional parameter for NewAppender.
//
// Applicable types: [CallerEncodeFunc], [ColorSetting], [PoolSizeLimit],
// [Config], [ConfigProvideFunc], [Environment], [FSOption],
// [Theme], [ThemeProvideFunc], [ThemeEnvironmentRef], [ThemeRef],
// [FlattenObjectsSetting], [TimestampEncodeFunc], [TimeValueEncodeFunc],
// [DurationEncodeFunc], [ErrorEncodeFunc].
type AppenderOption interface {
	toAppenderOptions(*appenderOptions)
}

// ---

// EncoderOption is an optional parameter for NewEncoder.
//
// Applicable types: [CallerEncodeFunc], [ColorSetting], [PoolSizeLimit],
// [Config], [ConfigProvideFunc], [Environment], [FSOption],
// [Theme], [ThemeProvideFunc], [ThemeEnvironmentRef], [ThemeRef],
// [FlattenObjectsSetting], [TimestampEncodeFunc], [TimeValueEncodeFunc],
// [DurationEncodeFunc], [ErrorEncodeFunc].
type EncoderOption interface {
	AppenderOption
	toEncoderOptions(*encoderOptions)
}

// ---

// PoolSizeLimit limits size of the pool of pre-allocated entry encoders.
type PoolSizeLimit int

func (v PoolSizeLimit) toEncoderOptions(o *encoderOptions) {
	o.poolSizeLimit = v
}

func (v PoolSizeLimit) toAppenderOptions(o *appenderOptions) {
	o.poolSizeLimit = v
}

// ---

func defaultAppenderOptions() appenderOptions {
	return appenderOptions{
		defaultEncoderOptions(),
	}
}

type appenderOptions struct {
	encoderOptions
}

func (o appenderOptions) With(other []AppenderOption) appenderOptions {
	for _, oo := range other {
		oo.toAppenderOptions(&o)
	}

	return o
}

// ---

func defaultEncoderOptions() encoderOptions {
	return encoderOptions{
		domain:         defaultDomain(),
		provideConfig:  []ConfigProvideFunc{ConfigFromEnvironment()},
		provideTheme:   []ThemeProvideFunc{ThemeFromEnvironment().fn()},
		poolSizeLimit:  8,
		flattenObjects: true,
	}
}

type encoderOptions struct {
	domain
	color           ColorSetting
	provideConfig   []ConfigProvideFunc
	provideTheme    []ThemeProvideFunc
	encodeCaller    CallerEncodeFunc
	encodeError     ErrorEncodeFunc
	encodeTimestamp TimestampEncodeFunc
	encodeTimeValue TimeValueEncodeFunc
	encodeDuration  DurationEncodeFunc
	poolSizeLimit   PoolSizeLimit
	flattenObjects  bool
}

func (o encoderOptions) With(other []EncoderOption) encoderOptions {
	for _, oo := range other {
		oo.toEncoderOptions(&o)
	}

	return o
}

// ---

var (
	_ AppenderOption = ColorAlways
	_ EncoderOption  = ColorAlways
	_ AppenderOption = PoolSizeLimit(0)
	_ EncoderOption  = PoolSizeLimit(0)
)
