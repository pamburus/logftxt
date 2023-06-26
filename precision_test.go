package logftxt_test

import (
	"testing"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt"
)

func TestPrecision(tt *testing.T) {
	t := tst.New(tt)

	t.Run("MarshalText", func(t tst.Test) {
		t.Expect(logftxt.Precision(3).MarshalText()).ToSucceed().AndResult().ToEqual([]byte("3"))
		t.Expect(logftxt.Precision(0).MarshalText()).ToSucceed().AndResult().ToEqual([]byte("0"))
		t.Expect(logftxt.PrecisionAuto.MarshalText()).ToSucceed().AndResult().ToEqual([]byte("auto"))
	})

	t.Run("UnmarshalText", func(t tst.Test) {
		t.Run("Success", func(t tst.Test) {
			var p logftxt.Precision
			t.Expect(p.UnmarshalText([]byte("auto"))).ToSucceed()
			t.Expect(p).ToEqual(logftxt.PrecisionAuto)

			t.Expect(p.UnmarshalText([]byte("3"))).ToSucceed()
			t.Expect(p).ToEqual(logftxt.Precision(3))

			t.Expect(p.UnmarshalText([]byte("0"))).ToSucceed()
			t.Expect(p).ToEqual(logftxt.Precision(0))
		})

		t.Run("Failure", func(t tst.Test) {
			var p logftxt.Precision
			t.Expect(p.UnmarshalText([]byte("bad"))).ToFail()
		})
	})
}
