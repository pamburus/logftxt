package logftxt_test

import (
	"bytes"
	"testing"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt"
)

func TestTheme(tt *testing.T) {
	t := tst.New(tt)

	t.Run("BuiltIn", func(t tst.Test) {
		t.Run("Default", func(t tst.Test) {
			_ = logftxt.DefaultTheme()
			t.Expect(logftxt.NewThemeRef("@default").Load()).ToSucceed().AndResult().ToNot(tst.BeZero())
		})

		t.Run("Others", func(t tst.Test) {
			themes, err := logftxt.ListBuiltInThemes()
			t.Expect(err).ToNot(tst.HaveOccurred())

			for _, theme := range themes {
				t.Expect(logftxt.LoadBuiltInTheme(theme)).ToSucceed()
			}
		})

		t.Run("NonExistent", func(t tst.Test) {
			t.Expect(logftxt.LoadBuiltInTheme("non-existent")).ToFail()
			t.Expect(logftxt.LoadTheme("./non-existent")).ToFail()
		})

		t.Run("Invalid", func(t tst.Test) {
			t.Expect(logftxt.ReadTheme(bytes.NewBufferString("asdasdh"))).ToFail()
		})
	})

	t.Run("Ref", func(t tst.Test) {
		t.Run("String", func(t tst.Test) {
			t.Expect(logftxt.NewThemeRef("@aa").String()).ToEqual("@aa")
		})
		t.Run("MarshalText", func(t tst.Test) {
			t.Expect(logftxt.NewThemeRef("@aa").MarshalText()).ToSucceed().AndResult().ToEqual([]byte("@aa"))
		})
	})
}
