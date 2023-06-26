// Package env provides accessors to the known environment variables.
package env

import (
	"os"
	"strings"
)

// ColorSetting checks environment variables for variables LOGFTXT_COLOR and NO_COLOR.
//
// All command-line software which outputs text with ANSI color added should
// check for the presence of a NO_COLOR environment variable that, when
// present (regardless of its value), prevents the addition of ANSI color.
func ColorSetting(lookup LookupFunc) Color {
	if setting, ok := lookup(envColorSetting); ok {
		switch {
		case strings.EqualFold(setting, string(ColorAuto)):
			return ColorAuto
		case strings.EqualFold(setting, string(ColorAlways)):
			return ColorAlways
		case strings.EqualFold(setting, string(ColorNever)):
			return ColorNever
		}
	}

	if isSome(lookup(envNoColor)) {
		return ColorNever
	}

	return ColorAuto
}

// Config returns configuration file path specified via environment variables.
func Config(lookup LookupFunc) (string, bool) {
	return lookup(envConfig)
}

// Theme returns theme name specified via environment variables.
func Theme(lookup LookupFunc) (string, bool) {
	return lookup(envTheme)
}

// ---

// Color is a color setting that can be specified via environment variables.
type Color string

// Valid values for Color setting.
const (
	ColorAuto   Color = "auto"
	ColorAlways Color = "always"
	ColorNever  Color = "never"
)

// ---

// Unset removes all known environment variables from the current process.
// Can be useful for unit tests to avoid dependency on environment.
func Unset() {
	for _, v := range []string{envNoColor, envColorSetting, envConfig, envTheme} {
		err := os.Unsetenv(v)
		if err != nil {
			panic(err)
		}
	}
}

// ---

// LookupFunc is an environment variable lookup function.
type LookupFunc = func(string) (string, bool)

// ---

func isSome(_ string, ok bool) bool {
	return ok
}

// ---

const (
	envNoColor      = "NO_COLOR"
	envColorSetting = "LOGFTXT_COLOR"
	envConfig       = "LOGFTXT_CONFIG"
	envTheme        = "LOGFTXT_THEME"
)
