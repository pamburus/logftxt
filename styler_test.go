package logftxt

import (
	"testing"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-ansi-esc/sgr"
	"github.com/pamburus/go-tst/tst"
)

func TestStyler(tt *testing.T) {
	t := tst.New(tt)

	t.Run("StylePatch", func(t tst.Test) {
		buf := logf.NewBuffer()
		s := newStyler()
		s.Use(stylePatch{Modes: [3]sgr.ModeSet{sgr.Faint.ModeSet()}, HasModes: true}, buf, func() {
			buf.AppendByte('a')
			s.Use(stylePatch{HasModes: true}, buf, func() {
				buf.AppendByte('b')
			})
		})
		t.Expect(string(buf.Data)).ToEqual("\x1b[2mab\x1b[0m")
	})

	t.Run("Style", func(t tst.Test) {
		initial := style{Background: sgr.SetBackgroundColor(sgr.Blue)}
		expected := initial
		expected.Background = sgr.SetBackgroundColor(sgr.Red)
		final := initial.UpdatedBy(stylePatch{Background: sgr.SetBackgroundColor(sgr.Red)})
		t.Expect(final).ToEqual(expected)
	})
}
