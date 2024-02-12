// Package formatting provides formatting configuration section for slogtxt theme configuration.
package formatting

import "github.com/pamburus/go-ansi-esc/sgr"

// ---

// Item is an item inside formatting section.
type Item struct {
	Outer     Format     `yaml:"outer"`
	Inner     Format     `yaml:"inner"`
	Separator StyledText `yaml:"separator,omitempty"`
	Text      string     `yaml:"text,omitempty"`
}

// UpdatedBy returns a copy of i updated by other.
func (i Item) UpdatedBy(other Item) Item {
	i.Outer = i.Outer.UpdatedBy(other.Outer)
	i.Inner = i.Inner.UpdatedBy(other.Inner)
	if other.Separator.Text != "" {
		i.Separator.Text = other.Separator.Text
	}
	if other.Separator.Style != nil {
		i.Separator.Style = other.Separator.Style
	}
	if other.Text != "" {
		i.Text = other.Text
	}

	return i
}

// ---

// StyledText is a text with style.
type StyledText struct {
	Text  string `yaml:"text,omitempty"`
	Style *Style `yaml:"style,omitempty"`
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
}

// ---

// Style includes background color, foreground color and a set of modes.
// Modes overwrites current set of modes during style rendering.
// So, explicitly specifying empty list of modes will disable all currently enabled modes.
type Style struct {
	Background sgr.Color     `yaml:"background,omitempty"`
	Foreground sgr.Color     `yaml:"foreground,omitempty"`
	Modes      ModePatchList `yaml:"modes,omitempty"`
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

// ---

// ModePatchList is a list of mode patches.
type ModePatchList []ModePatch

// Sets returns an array of mode sets for each possible sgr.ModeAction.
func (l ModePatchList) Sets() [4]sgr.ModeSet {
	var sets [4]sgr.ModeSet
	for _, p := range l {
		sets[p.Action] |= p.Mode.ModeSet()
	}

	return sets
}

// ---

// ModePatch is a patch for a mode.
type ModePatch struct {
	Mode   sgr.Mode
	Action sgr.ModeAction
}

// MarshalText returns a text representation of p.
func (p ModePatch) MarshalText() ([]byte, error) {
	mode, err := p.Mode.MarshalText()
	if err != nil {
		return nil, err
	}

	switch p.Action {
	case sgr.ModeReplace:
		return mode, nil
	case sgr.ModeAdd:
		return append([]byte("+"), mode...), nil
	case sgr.ModeRemove:
		return append([]byte("-"), mode...), nil
	case sgr.ModeToggle:
		return append([]byte("^"), mode...), nil
	default:
		panic("unknown mode action")
	}
}

// UnmarshalText parses text and stores the result in p.
func (p *ModePatch) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	var action sgr.ModeAction
	switch text[0] {
	case '+':
		action = sgr.ModeAdd
		text = text[1:]
	case '-':
		action = sgr.ModeRemove
		text = text[1:]
	case '^':
		action = sgr.ModeToggle
		text = text[1:]
	default:
		action = sgr.ModeReplace
	}

	var mode sgr.Mode
	err := mode.UnmarshalText(text)
	if err != nil {
		return err
	}

	p.Mode = mode
	p.Action = action

	return nil
}
