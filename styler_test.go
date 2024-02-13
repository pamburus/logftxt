package logftxt

import (
	"testing"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-ansi-esc/sgr"
	"github.com/pamburus/go-tst/tst"
)

func TestStyler(tt *testing.T) {
	t := tst.New(tt)

	buf := logf.NewBuffer()
	s := newStyler()
	s.Use(stylePatch{Modes: [3]sgr.ModeSet{sgr.Faint.ModeSet()}, HasModes: true}, buf, func() {
		buf.AppendByte('a')
		s.Use(stylePatch{HasModes: true}, buf, func() {
			buf.AppendByte('b')
		})
	})

	t.Expect(string(buf.Data)).ToEqual("\x1b[2mab\x1b[0m")
}
