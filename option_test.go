package logftxt

import (
	"testing"
	"time"

	"github.com/pamburus/go-tst/tst"
)

func TestOptions(tt *testing.T) {
	t := tst.New(tt)

	ao := func(options ...AppenderOption) appenderOptions {
		return appenderOptions{}.With(options)
	}
	eo := func(options ...EncoderOption) encoderOptions {
		return encoderOptions{}.With(options)
	}

	t.Expect(ao(ColorAlways).color).ToNot(tst.BeZero())
	t.Expect(ao(ColorAlways).color).ToNot(tst.BeZero())
	t.Expect(ao(TimeLayout("as").Timestamp()).encodeTimestamp).ToNot(tst.BeZero())
	t.Expect(eo(TimeLayout("as").Timestamp()).encodeTimestamp).ToNot(tst.BeZero())
	t.Expect(ao(TimeLayout("as").TimeValue()).encodeTimeValue).ToNot(tst.BeZero())
	t.Expect(eo(TimeLayout("as").TimeValue()).encodeTimeValue).ToNot(tst.BeZero())
	t.Expect(ao(PoolSizeLimit(1)).poolSizeLimit).ToNot(tst.BeZero())
	t.Expect(eo(PoolSizeLimit(1)).poolSizeLimit).ToNot(tst.BeZero())
	t.Expect(ao(DurationAsHMS(PrecisionAuto)).encodeDuration).ToNot(tst.BeZero())
	t.Expect(eo(DurationAsHMS(Precision(2))).encodeDuration).ToNot(tst.BeZero())
	t.Expect(ao(ErrorLong()).encodeError).ToNot(tst.BeZero())
	t.Expect(eo(ErrorLong()).encodeError).ToNot(tst.BeZero())
	t.Expect(ao(&Config{}).provideConfig).ToNot(tst.BeZero())
	t.Expect(ao(&Config{}).provideConfig).ToNot(tst.BeZero())
	t.Expect(ao(ConfigFromDefaultPath()).provideConfig).ToNot(tst.BeZero())
	t.Expect(ao(ConfigFromDefaultPath()).provideConfig).ToNot(tst.BeZero())
	t.Expect(ao(DefaultTheme()).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(DefaultTheme()).provideTheme).ToNot(tst.BeZero())
	t.Expect(NewThemeRef("").Load()).To(tst.BeZero(), tst.Equal(nil))
	t.Expect(ao(NewThemeRef("")).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(NewThemeRef("")).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(NewThemeRef("").fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(NewThemeRef("").fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(ThemeFromEnvironment()).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(ThemeFromEnvironment()).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(ThemeFromEnvironment().fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(ao(ThemeFromEnvironment().fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(ThemeFromEnvironment()).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(ThemeFromEnvironment()).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(ThemeFromEnvironment().fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(eo(ThemeFromEnvironment().fn()).provideTheme).ToNot(tst.BeZero())
	t.Expect(
		eo(
			ThemeFromEnvironment(),
		).provideTheme[0](
			domain{env: Environment(mockEnv{"LOGFTXT_THEME": "@default"}.lookup)},
		),
	).ToSucceed().AndResult().ToNot(tst.BeZero())
	t.Expect(defaultEncoderOptions()).ToNot(tst.BeZero())
	t.Expect(defaultAppenderOptions()).ToNot(tst.BeZero())
	t.Expect(DefaultDomain().Environment()).ToNotEqual(nil)
	t.Expect(DefaultDomain().FS()).ToNotEqual(nil)
	t.Expect(ao(WithFS(SystemFS())).fs).ToNot(tst.BeZero())
	t.Expect(eo(WithFS(SystemFS())).fs).ToNot(tst.BeZero())
}

func TestTimeLayout(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(
		string(TimeLayout(time.RFC3339Nano)(nil, time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC))),
	).ToEqual(
		"2020-01-02T03:04:05.000000006Z",
	)
}

// ---

type mockEnv map[string]string

func (e mockEnv) lookup(key string) (string, bool) {
	result, ok := e[key]

	return result, ok
}
