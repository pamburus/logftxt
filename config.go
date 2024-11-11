package logftxt

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/pamburus/logftxt/internal/pkg/env"
)

// LoadConfig loads configuration file with the given filename.
func LoadConfig(filename string, opts ...domainOption) (*Config, error) {
	o := defaultDomain().with(opts)

	return loadConfig(filename, o.fs)
}

// ReadConfig loads configuration from the given reader.
func ReadConfig(reader io.Reader) (*Config, error) {
	fail := func(err error) (*Config, error) {
		return nil, err
	}

	var cfg Config

	err := yaml.NewDecoder(reader).Decode(&cfg)
	if err != nil {
		return fail(fmt.Errorf("failed to parse config: %w", err))
	}

	err = cfg.Validate()
	if err != nil {
		return fail(fmt.Errorf("config is invalid: %w", err))
	}

	return &cfg, nil
}

// DefaultConfig returns default built-in configuration.
func DefaultConfig() *Config {
	return embeddedConfigs.Default.Load()
}

// ---

// Config holds logftxt configuration.
type Config struct {
	Theme     ThemeRef `yaml:"theme"`
	Timestamp struct {
		Format string `yaml:"format"`
	} `yaml:"timestamp"`
	Caller struct {
		Format CallerFormat `yaml:"format"`
	} `yaml:"caller"`
	Values struct {
		Time struct {
			Format string `yaml:"format"`
		} `yaml:"time"`
		Duration struct {
			Format    DurationFormat `yaml:"format"`
			Precision Precision      `yaml:"precision"`
		} `yaml:"duration"`
		Error struct {
			Format ErrorFormat `yaml:"format"`
		} `yaml:"error"`
	} `yaml:"values"`
}

// Validate checks whether c is valid.
func (c Config) Validate() error {
	err := c.Caller.Format.Validate()
	if err != nil {
		return fmt.Errorf("caller format is invalid: %w", err)
	}

	err = c.Values.Duration.Format.Validate()
	if err != nil {
		return fmt.Errorf("duration format is invalid: %w", err)
	}

	err = c.Values.Error.Format.Validate()
	if err != nil {
		return fmt.Errorf("error format is invalid: %w", err)
	}

	return nil
}

func (c Config) toEncoderOptions(oo *encoderOptions) {
	oo.provideConfig = append(oo.provideConfig, c.fn())
}

func (c Config) toAppenderOptions(oo *appenderOptions) {
	oo.provideConfig = append(oo.provideConfig, c.fn())
}

func (c Config) fn() ConfigProvideFunc {
	return func(Domain) (*Config, error) {
		return &c, nil
	}
}

// ---

// ConfigFromEnvironment returns a ConfigProvideFunc that
// gets configuration filename from environment variable `LOGFTXT_CONFIG`.
// In case there is no such environment variable, `nil` is returned
// allowing to fallback to other configuration sources.
func ConfigFromEnvironment(opts ...domainOption) ConfigProvideFunc {
	return func(domain Domain) (*Config, error) {
		domain = NewDomain(domain, opts...)
		if configPath, ok := env.Config(domain.Environment()); ok {
			return loadConfig(configPath, domain.FS())
		}

		return nil, nil
	}
}

// ConfigFromDefaultPath returns a ConfigProvideFunc that
// load configuration from default path that is `~/.config/logftxt/config.yml`.
// In case there is no such file, `nil` is returned
// allowing to fallback to other configuration sources.
func ConfigFromDefaultPath(opts ...domainOption) ConfigProvideFunc {
	return func(domain Domain) (*Config, error) {
		return loadConfig("config.yml", NewDomain(domain, opts...).FS())
	}
}

// ---

// ConfigProvideFunc is a function that provides Config when called.
// ConfigProvideFunc can return `nil` configuration and `nil` error
// in case there is no any configuration found at the corresponding source.
type ConfigProvideFunc func(Domain) (*Config, error)

func (f ConfigProvideFunc) toEncoderOptions(oo *encoderOptions) {
	oo.provideConfig = append(oo.provideConfig, f)
}

func (f ConfigProvideFunc) toAppenderOptions(oo *appenderOptions) {
	oo.provideConfig = append(oo.provideConfig, f)
}

// ---

// Valid values for DurationFormat.
const (
	DurationFormatDefault DurationFormat = ""
	DurationFormatDynamic DurationFormat = "dynamic"
	DurationFormatSeconds DurationFormat = "seconds"
	DurationFormatHMS     DurationFormat = "hms"
)

// DurationFormat defines duration output format.
type DurationFormat string

// Validate checks whether v has a valid value.
func (v DurationFormat) Validate() error {
	switch v {
	case DurationFormatDefault:
	case DurationFormatDynamic:
	case DurationFormatSeconds:
	case DurationFormatHMS:
	default:
		return fmt.Errorf("unknown duration format %q", v)
	}

	return nil
}

// ---

// Valid values for CallerFormat.
const (
	CallerFormatDefault CallerFormat = ""
	CallerFormatShort   CallerFormat = "short"
	CallerFormatLong    CallerFormat = "long"
)

// CallerFormat defines duration output format.
type CallerFormat string

// Validate checks whether v has a valid value.
func (v CallerFormat) Validate() error {
	switch v {
	case CallerFormatDefault:
	case CallerFormatShort:
	case CallerFormatLong:
	default:
		return fmt.Errorf("unknown caller format %q", v)
	}

	return nil
}

// ---

// Valid values for ErrorFormat.
const (
	ErrorFormatDefault ErrorFormat = ""
	ErrorFormatShort   ErrorFormat = "short"
	ErrorFormatLong    ErrorFormat = "long"
)

// ErrorFormat defines duration output format.
type ErrorFormat string

// Validate checks whether v has a valid value.
func (v ErrorFormat) Validate() error {
	switch v {
	case ErrorFormatDefault:
	case ErrorFormatShort:
	case ErrorFormatLong:
	default:
		return fmt.Errorf("unknown error format %q", v)
	}

	return nil
}

// ---

func loadConfig(filename string, fileSystem FS) (*Config, error) {
	f, err := fileSystem.Open(filename) //nolint:gosec // it is ok to allow user to specify config file path
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrFileNotFound{filename, err}
		}

		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	defer f.Close() //nolint:errcheck // read-only

	return ReadConfig(f)
}

// ---

//go:embed assets/config.yml
var embeddedDefaultConfigBytes []byte

var embeddedConfigs = struct {
	Default embeddedConfigLoader
}{
	Default: embeddedConfigLoader{data: embeddedDefaultConfigBytes},
}

// ---

type embeddedConfigLoader struct {
	data   []byte
	config *Config
	once   sync.Once
}

func (t *embeddedConfigLoader) Load() *Config {
	t.once.Do(func() {
		config, err := ReadConfig(bytes.NewReader(t.data))
		if err != nil {
			panic(err)
		}

		t.config = config
	})

	return t.config
}

// ---

// ---

var (
	_ EncoderOption  = Config{}
	_ AppenderOption = Config{}
	_ EncoderOption  = ConfigProvideFunc(nil)
	_ AppenderOption = ConfigProvideFunc(nil)
)
