// Package formatting provides formatting configuration section for logftxt theme configuration.
package formatting

import "github.com/pamburus/go-ansi-esc/sgr"

// ---

// Item is an item inside formatting section.
type Item struct {
	Outer     Format `yaml:"outer"`
	Inner     Format `yaml:"inner"`
	Separator string `yaml:"separator,omitempty"`
	Text      string `yaml:"text,omitempty"`
}

// UpdatedBy returns a copy of i updated by other.
func (i Item) UpdatedBy(other Item) Item {
	i.Outer = i.Outer.UpdatedBy(other.Outer)
	i.Inner = i.Inner.UpdatedBy(other.Inner)
	if other.Separator != "" {
		i.Separator = other.Separator
	}
	if other.Text != "" {
		i.Text = other.Text
	}

	return i
}

// ---

// Format contains prefix, suffix and style, each is optional.
type Format struct {
	Prefix string `yaml:"prefix,omitempty"`
	Suffix string `yaml:"suffix,omitempty"`
	Style  Style  `yaml:"style"`
}

// UpdatedBy returns a copy of f updated by other.
func (f Format) UpdatedBy(other Format) Format {
	if other.Prefix != "" {
		f.Prefix = other.Prefix
	}
	if other.Suffix != "" {
		f.Suffix = other.Suffix
	}
	f.Style = f.Style.UpdatedBy(other.Style)

	return f
}

// ---

// Level is a log level formatting configuration.
type Level struct {
	All     Item `yaml:"all"`
	Debug   Item `yaml:"debug"`
	Info    Item `yaml:"info"`
	Warning Item `yaml:"warning"`
	Error   Item `yaml:"error"`
	Unknown Item `yaml:"unknown"`
}

// ---

// Style includes background color, foreground color and a set of modes.
// Modes overwrites current set of modes during style rendering.
// So, explicitly specifying empty list of modes will disable all currently enabled modes.
type Style struct {
	Background sgr.Color    `yaml:"background,omitempty"`
	Foreground sgr.Color    `yaml:"foreground,omitempty"`
	Modes      sgr.ModeList `yaml:"modes,omitempty"`
}

// UpdatedBy returns a copy of s updated by other.
func (s Style) UpdatedBy(other Style) Style {
	if !other.Background.IsZero() {
		s.Background = other.Background
	}
	if !other.Foreground.IsZero() {
		s.Foreground = other.Foreground
	}
	if other.Modes != nil {
		s.Modes = other.Modes
	}

	return s
}
