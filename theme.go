package logftxt

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-ansi-esc/sgr"
	"github.com/pamburus/logftxt/internal/pkg/env"
	"github.com/pamburus/logftxt/internal/pkg/themecfg"
	"github.com/pamburus/logftxt/internal/pkg/themecfg/formatting"
)

// LoadTheme loads theme from a file defined by the given filename.
//
// If the filename is relative and does not start with `./` or `../` then
// the file will be searched in `~/.config/logftxt/themes` folder.
//
// To search file relatively current working directory instead, add explicit `./` or `../` prefix.
func LoadTheme(filename string, opts ...domainOption) (*Theme, error) {
	o := defaultDomain().with(opts)

	return loadTheme(filename, o.fs)
}

// ReadTheme reads theme configuration from the given reader.
func ReadTheme(reader io.Reader) (*Theme, error) {
	cfg, err := themecfg.Load(reader)
	if err != nil {
		return nil, err
	}

	return newTheme(cfg), nil
}

// DefaultTheme returns default built-in theme.
func DefaultTheme() *Theme {
	defaultThemeOnce.Do(func() {
		var err error
		defaultTheme, err = LoadBuiltInTheme(defaultThemeName)
		if err != nil {
			panic(err)
		}
	})

	return defaultTheme
}

// LoadBuiltInTheme loads named built-in theme.
func LoadBuiltInTheme(name string) (*Theme, error) {
	f, err := embeddedThemes.Open(path.Join("assets/theme", name+".yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("built-in theme %q not found", name)
		}

		return nil, fmt.Errorf("failed to load built-in theme %q: %v", name, err)
	}

	return ReadTheme(f)
}

// ListBuiltInThemes returns list of names of all built-in themes.
func ListBuiltInThemes() ([]string, error) {
	matches, err := fs.Glob(embeddedThemes, "assets/theme/*.yml")
	if err != nil {
		return nil, err
	}

	for i := range matches {
		matches[i] = strings.TrimSuffix(path.Base(matches[i]), ".yml")
	}

	return matches, nil
}

// ---

// Theme holds formatting and styling settings.
type Theme struct {
	items []item
	fmt   fmtItems
}

func (t *Theme) toEncoderOptions(o *encoderOptions) {
	o.provideTheme = append(o.provideTheme, t.fn())
}

func (t *Theme) toAppenderOptions(o *appenderOptions) {
	o.provideTheme = append(o.provideTheme, t.fn())
}

func (t *Theme) fn() ThemeProvideFunc {
	return func(Domain) (*Theme, error) {
		return t, nil
	}
}

// ---

// ThemeProvideFunc is a function that provides Theme when called.
//
// ThemeProvideFunc can return `nil` theme and `nil` error
// meaning there is nothing to provide from its source.
type ThemeProvideFunc func(Domain) (*Theme, error)

func (f ThemeProvideFunc) toEncoderOptions(o *encoderOptions) {
	o.provideTheme = append(o.provideTheme, f)
}

func (f ThemeProvideFunc) toAppenderOptions(o *appenderOptions) {
	o.provideTheme = append(o.provideTheme, f)
}

// ---

// ThemeFromEnvironment returns a ThemeRef that references some theme as defined by the environment variables.
func ThemeFromEnvironment(opts ...domainOption) ThemeEnvironmentRef {
	return ThemeEnvironmentRef{opts}
}

// ---

// ThemeEnvironmentRef is an option that requests to load theme specified by environment variables.
type ThemeEnvironmentRef struct {
	opts []domainOption
}

func (v ThemeEnvironmentRef) toEncoderOptions(o *encoderOptions) {
	o.provideTheme = append(o.provideTheme, v.fn())
}

func (v ThemeEnvironmentRef) toAppenderOptions(o *appenderOptions) {
	o.provideTheme = append(o.provideTheme, v.fn())
}

func (v ThemeEnvironmentRef) fn() ThemeProvideFunc {
	return func(domain Domain) (*Theme, error) {
		domain = NewDomain(domain, v.opts...)

		if v, ok := env.Theme(domain.Environment()); ok {
			return NewThemeRef(v, WithFS(domain.FS())).Load()
		}

		return nil, nil
	}
}

// ---

// NewThemeRef constructs a new theme reference with the given name and options.
func NewThemeRef(name string, opts ...domainOption) ThemeRef {
	return ThemeRef{name, opts}
}

// ThemeRef is a reference to a built-in or external theme.
// Built-in theme is referenced by '@' prefix following built-in theme name.
// External theme is referenced by specifying an absolute or relative path to a theme configuration file.
// Relative path starting with `./` or `../` will request locating the file relatively to the current working directory.
// Relative path not starting with `./` or `../` will request locating the file relatively to `~/.config/logftxt/themes` directory.
type ThemeRef struct {
	name    string
	options []domainOption
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ThemeRef) MarshalText() ([]byte, error) {
	return []byte(v.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ThemeRef) UnmarshalText(text []byte) error {
	v.name = string(text)

	return nil
}

// String returns theme reference name.
func (v ThemeRef) String() string {
	return v.name
}

// Load loads the referenced theme.
func (v ThemeRef) Load() (*Theme, error) {
	if v.name == "" {
		return nil, nil
	}

	if v.name[0] == '@' {
		return LoadBuiltInTheme(v.name[1:])
	}

	return loadTheme(v.name, defaultDomain().with(v.options).fs)
}

func (v ThemeRef) toEncoderOptions(o *encoderOptions) {
	o.provideTheme = append(o.provideTheme, v.fn())
}

func (v ThemeRef) toAppenderOptions(o *appenderOptions) {
	o.provideTheme = append(o.provideTheme, v.fn())
}

func (v ThemeRef) fn() ThemeProvideFunc {
	return ThemeProvideFunc(func(Domain) (*Theme, error) {
		return v.Load()
	})
}

// ---

func loadTheme(filename string, fs FS) (*Theme, error) {
	f, err := fs.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return ReadTheme(f)
}

// ---

type item interface {
	encode(*entryEncoder)
}

// ---

type itemTimestamp struct{}

func (*itemTimestamp) encode(e *entryEncoder) {
	e.theme.fmt.Timestamp.encode(e, func() {
		e.appendTimestamp(e.entry.Time)
	})
}

// ---

type itemLevel struct{}

func (*itemLevel) encode(e *entryEncoder) {
	// assert that logf.LevelDebug value is greater of equal than logf.LevelError value
	_ = [logf.LevelDebug - logf.LevelError]struct{}{}

	i := e.entry.Level
	if i < logf.LevelError {
		i = logf.LevelError
	}
	if i > logf.LevelDebug {
		i = logf.LevelDebug
	}

	level := e.theme.fmt.Level[i]

	level.encode(e, func() {
		e.buf.AppendString(level.text)
	})
}

// ---

type itemLogger struct{}

func (*itemLogger) encode(e *entryEncoder) {
	if e.entry.LoggerName != "" {
		e.theme.fmt.Logger.encode(e, func() {
			e.buf.AppendString(e.entry.LoggerName)
		})
	}
}

// ---

type itemMessage struct{}

func (*itemMessage) encode(e *entryEncoder) {
	if e.entry.Text != "" {
		e.theme.fmt.Message.encode(e, func() {
			e.buf.AppendString(e.entry.Text)
		})
	}
}

// ---

type itemCaller struct{}

func (*itemCaller) encode(e *entryEncoder) {
	if e.entry.Caller.Specified {
		e.theme.fmt.Caller.encode(e, func() {
			e.appendCaller(e.entry.Caller)
		})
	}
}

// ---

type itemFields struct{}

func (*itemFields) encode(e *entryEncoder) {
	// Logger's fields.
	if bytes, ok := e.cache.Get(e.entry.LoggerID); ok {
		e.buf.AppendBytes(bytes)
	} else {
		le := e.buf.Len()
		for _, field := range e.entry.DerivedFields {
			e.appendSeparator()
			field.Accept(e)
		}

		if n := e.buf.Len() - le; n != 0 {
			bf := make([]byte, n)
			copy(bf, e.buf.Data[le:])
			e.cache.Set(e.entry.LoggerID, bf)
		}
	}

	// Entry's fields.
	for _, field := range e.entry.Fields {
		e.appendSeparator()
		field.Accept(e)
	}
}

// ---

type fmtItem struct {
	outer     format
	inner     format
	separator styledText
	text      string
}

func (i *fmtItem) encode(e *entryEncoder, encodeInner func()) {
	e.styler.Use(i.outer.style, e.buf, func() {
		e.buf.AppendString(i.outer.prefix)
		e.styler.Use(i.inner.style, e.buf, encodeInner)
		e.buf.AppendString(i.outer.suffix)
	})
}

// ---

type format struct {
	prefix string
	suffix string
	style  stylePatch
}

// ---

type styledText struct {
	text  string
	style stylePatch
}

func (t styledText) encode(e *entryEncoder) {
	e.styler.Use(t.style, e.buf, func() {
		e.buf.AppendString(t.text)
	})
}

// ---

type fmtItems struct {
	Timestamp fmtItem
	Level     [4]fmtItem
	Logger    fmtItem
	Message   fmtItem
	Field     fmtItem
	Key       fmtItem
	Caller    fmtItem
	Array     fmtItem
	Object    fmtItem
	String    fmtItem
	Number    fmtItem
	Boolean   fmtItem
	Time      fmtItem
	Duration  fmtItem
	Null      fmtItem
	Error     fmtItem
}

// ---

func newTheme(cfg *themecfg.Theme) *Theme {
	var items []item

	for _, it := range cfg.Items {
		if it, ok := newItem(it); ok {
			items = append(items, it)
		}
	}

	return &Theme{
		items,
		fmtItems{
			newFmtItem(cfg.Formatting.Timestamp),
			[4]fmtItem{
				logf.LevelDebug: newFmtItem(cfg.Formatting.Level.All.UpdatedBy(cfg.Formatting.Level.Debug)),
				logf.LevelInfo:  newFmtItem(cfg.Formatting.Level.All.UpdatedBy(cfg.Formatting.Level.Info)),
				logf.LevelWarn:  newFmtItem(cfg.Formatting.Level.All.UpdatedBy(cfg.Formatting.Level.Warning)),
				logf.LevelError: newFmtItem(cfg.Formatting.Level.All.UpdatedBy(cfg.Formatting.Level.Error)),
			},
			newFmtItem(cfg.Formatting.Logger),
			newFmtItem(cfg.Formatting.Message),
			newFmtItem(cfg.Formatting.Field),
			newFmtItem(cfg.Formatting.Key),
			newFmtItem(cfg.Formatting.Caller),
			newFmtItem(cfg.Formatting.Types.Array),
			newFmtItem(cfg.Formatting.Types.Object),
			newFmtItem(cfg.Formatting.Types.String),
			newFmtItem(cfg.Formatting.Types.Number),
			newFmtItem(cfg.Formatting.Types.Boolean),
			newFmtItem(cfg.Formatting.Types.Time),
			newFmtItem(cfg.Formatting.Types.Duration),
			newFmtItem(cfg.Formatting.Types.Null),
			newFmtItem(cfg.Formatting.Types.Error),
		},
	}
}

func newItem(it themecfg.Item) (item, bool) {
	switch it {
	case themecfg.ItemTimestamp:
		return &itemTimestamp{}, true
	case themecfg.ItemLevel:
		return &itemLevel{}, true
	case themecfg.ItemLogger:
		return &itemLogger{}, true
	case themecfg.ItemMessage:
		return &itemMessage{}, true
	case themecfg.ItemFields:
		return &itemFields{}, true
	case themecfg.ItemCaller:
		return &itemCaller{}, true
	default:
		return nil, false
	}
}

func newFmtItem(it formatting.Item) fmtItem {
	return fmtItem{
		outer:     newFormat(it.Outer),
		inner:     newFormat(it.Inner),
		separator: newStyledText(it.Separator),
		text:      it.Text,
	}
}

func newFormat(f formatting.Format) format {
	return format{
		prefix: f.Prefix,
		suffix: f.Suffix,
		style:  newStylePatch(f.Style),
	}
}

func newStylePatch(s themecfg.Style) stylePatch {
	return stylePatch{
		sgr.SetBackgroundColor(s.Background),
		sgr.SetForegroundColor(s.Foreground),
		s.Modes.Sets(),
		len(s.Modes) != 0,
		len(s.Modes) == 0 && s.Background.IsZero() && s.Foreground.IsZero(),
	}
}

func newStyledText(f formatting.StyledText) styledText {
	result := styledText{
		text: f.Text,
	}
	if !f.Style.IsZero() {
		result.style = newStylePatch(f.Style)
	}

	return result
}

// ---

//go:embed assets/theme/*.yml
var embeddedThemes embed.FS

// ---

const defaultThemeName = "default"

var defaultThemeOnce sync.Once
var defaultTheme *Theme

// ---

var (
	_ item = (*itemTimestamp)(nil)
	_ item = (*itemLevel)(nil)
	_ item = (*itemLogger)(nil)
	_ item = (*itemMessage)(nil)
	_ item = (*itemFields)(nil)
	_ item = (*itemCaller)(nil)
)
