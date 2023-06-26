package logftxt

import (
	"testing"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt/internal/pkg/pathx"
)

func TestAppender(tt *testing.T) {
	t := tst.New(tt)

	config, err := LoadConfig("./encoder_test.config.yml")
	t.Expect(err).ToNot(tst.HaveOccurred())

	theme := NewThemeRef(pathx.ExplicitlyRelative("encoder_test.theme.yml"))

	envColor := func(value string) Environment {
		return func(name string) (string, bool) {
			switch name {
			case "LOGFTXT_COLOR":
				return value, true
			default:
				return "", false
			}
		}
	}

	t.Run("Env", func(t tst.Test) {
		t.Run("Color", func(t tst.Test) {
			t.Run("Always", func(t tst.Test) {
				buf := logf.NewBuffer()
				appender := NewAppender(buf, envColor("always"), theme, config)
				t.Expect(appender.Append(logf.Entry{Text: "msg"})).ToSucceed()
				t.Expect(appender.Flush()).ToSucceed()
				t.Expect(buf.String()).ToEqual("\x1b[2mJan  1 00:00:00.000\x1b[0m \x1b[91;7m|ERR|\x1b[0m \x1b[1mmsg\x1b[0m\n")
			})
			t.Run("Never", func(t tst.Test) {
				buf := logf.NewBuffer()
				appender := NewAppender(buf, envColor("never"), theme, config)
				t.Expect(appender.Append(logf.Entry{Text: "msg"})).ToSucceed()
				t.Expect(appender.Flush()).ToSucceed()
				t.Expect(buf.String()).ToEqual("Jan  1 00:00:00.000 |ERR| msg\n")
			})
		})
	})
}
