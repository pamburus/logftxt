// Package themecfg defines configuration file format for a logftxt theme.
package themecfg

import (
	_ "embed" // needed for embedded assets
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/pamburus/logftxt/internal/pkg/themecfg/formatting"
)

// ---

// Load loads Theme from the given reader.
func Load(reader io.Reader) (*Theme, error) {
	fail := func(err error) (*Theme, error) {
		return nil, err
	}

	var content struct {
		Theme *Theme `yaml:"theme"`
	}

	err := yaml.NewDecoder(reader).Decode(&content)
	if err != nil {
		return fail(fmt.Errorf("failed to parse theme: %v", err))
	}

	if content.Theme == nil {
		return fail(errors.New("invalid theme format"))
	}

	if content.Theme.Version != "1.0" {
		return fail(errors.New("unsupported theme version"))
	}

	err = content.Theme.Validate()
	if err != nil {
		return fail(fmt.Errorf("theme is invalid: %v", err))
	}

	return content.Theme, nil
}

// ---

// Theme contains theme configuration that can be described in a YAML file.
type Theme struct {
	Version    string     `yaml:"version"`
	Items      []Item     `yaml:"items"`
	Settings   Settings   `yaml:"settings"`
	Formatting Formatting `yaml:"formatting"`
}

// Validate check that t is valid.
func (t *Theme) Validate() error {
	if len(t.Items) == 0 {
		return errors.New("`items` should not be empty")
	}

	for i, item := range t.Items {
		err := item.Validate()
		if err != nil {
			return fmt.Errorf("`items.%d` is invalid: %v", i, err)
		}
	}

	return nil
}

// ---

// Valid values for Item.
const (
	ItemTimestamp Item = "timestamp"
	ItemLevel     Item = "level"
	ItemLogger    Item = "logger"
	ItemMessage   Item = "message"
	ItemFields    Item = "fields"
	ItemCaller    Item = "caller"
)

// Item defines and item that can go as part of log message output.
type Item string

// Validate checks if i has a valid value.
func (i Item) Validate() error {
	switch i {
	case ItemTimestamp:
	case ItemLevel:
	case ItemLogger:
	case ItemMessage:
	case ItemFields:
	case ItemCaller:
	default:
		return fmt.Errorf("invalid value %q", i)
	}

	return nil
}

// ---

// Settings is a settings configuration section.
type Settings struct {
	TimeFormat string `yaml:"time-format"`
}

// ---

// Formatting is a formatting configuration section.
type Formatting struct {
	Timestamp formatting.Item  `yaml:"timestamp"`
	Level     formatting.Level `yaml:"level"`
	Logger    formatting.Item  `yaml:"logger"`
	Message   formatting.Item  `yaml:"message"`
	Field     formatting.Item  `yaml:"field"`
	Key       formatting.Item  `yaml:"key"`
	Caller    formatting.Item  `yaml:"caller"`
	Types     FormattingTypes  `yaml:"types"`
}

// ---

// FormattingTypes is a formatting.types configuration section.
type FormattingTypes struct {
	Array    formatting.Item `yaml:"array"`
	Object   formatting.Item `yaml:"object"`
	String   formatting.Item `yaml:"string"`
	Quotes   Style           `yaml:"quotes"`
	Special  Style           `yaml:"special"`
	Number   formatting.Item `yaml:"number"`
	Boolean  formatting.Item `yaml:"boolean"`
	Time     formatting.Item `yaml:"time"`
	Duration formatting.Item `yaml:"duration"`
	Null     formatting.Item `yaml:"null"`
	Error    formatting.Item `yaml:"error"`
}
