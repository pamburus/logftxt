package logftxt

import (
	"fmt"
	"testing"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-tst/tst"
)

func TestCaller(tt *testing.T) {
	t := tst.New(tt)

	const shortFilename = "logftxt/caller_test.go"
	caller := logf.NewEntryCaller(0)

	t.Run("Short", func(t tst.Test) {
		t.Expect(
			string(CallerShort()(nil, caller)),
		).ToEqual(
			fmt.Sprintf("%s:%d", shortFilename, caller.Line),
		)
	})

	t.Run("Long", func(t tst.Test) {
		t.Expect(
			string(CallerLong()(nil, caller)),
		).ToEqual(
			fmt.Sprintf("%s:%d", caller.File, caller.Line),
		)
	})

	t.Run("Options", func(t tst.Test) {
		f := CallerShort()
		t.Expect(encoderOptions{}.With([]EncoderOption{f}).encodeCaller).ToNot(tst.BeZero())
		t.Expect(appenderOptions{}.With([]AppenderOption{f}).encodeCaller).ToNot(tst.BeZero())
	})
}
