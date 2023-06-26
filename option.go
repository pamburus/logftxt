package logftxt

import "os"

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
	switch s {
	case ColorAlways:
		o.color = true
	case ColorNever:
		o.color = false
	}
}

// ---

// AppenderOption is an optional parameter for NewAppender.
type AppenderOption interface {
	toAppenderOptions(*appenderOptions)
}

// ---

// EncoderOption is an optional parameter for NewEncoder.
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
		ColorAuto,
		os.LookupEnv,
		defaultEncoderOptions(),
	}
}

type appenderOptions struct {
	color ColorSetting
	env   Environment
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
		provideConfig: []ConfigProvideFunc{ConfigFromEnvironment()},
		provideTheme:  []ThemeProvideFunc{ThemeFromEnvironment().fn()},
		poolSizeLimit: 8,
	}
}

type encoderOptions struct {
	provideConfig   []ConfigProvideFunc
	provideTheme    []ThemeProvideFunc
	color           bool
	encodeCaller    CallerEncodeFunc
	encodeError     ErrorEncodeFunc
	encodeTimestamp TimestampEncodeFunc
	encodeTimeValue TimeValueEncodeFunc
	encodeDuration  DurationEncodeFunc
	poolSizeLimit   PoolSizeLimit
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
