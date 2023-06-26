package logftxt_test

import (
	"testing"
	"time"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt"
)

func TestDuration(tt *testing.T) {
	t := tst.New(tt)

	t.Run("AsText", func(t tst.Test) {
		f := logftxt.DurationAsText()
		format := func(duration time.Duration) string {
			return string(f(nil, duration))
		}

		t.Expect(format(time.Second)).ToEqual("1s")
		t.Expect(format(time.Millisecond)).ToEqual("1ms")
		t.Expect(format(time.Microsecond)).ToEqual("1µs")
		t.Expect(format(time.Nanosecond)).ToEqual("1ns")
		t.Expect(format(42 * time.Minute)).ToEqual("42m0s")
	})

	t.Run("AsSeconds", func(t tst.Test) {
		t.Run("Auto", func(t tst.Test) {
			f := logftxt.DurationAsSeconds()
			format := func(duration time.Duration) string {
				return string(f(nil, duration))
			}

			t.Expect(format(time.Second)).ToEqual("1")
			t.Expect(format(time.Millisecond)).ToEqual("0.001")
			t.Expect(format(time.Microsecond)).ToEqual("0.000001")
			t.Expect(format(time.Nanosecond)).ToEqual("0.000000001")
			t.Expect(format(42 * time.Minute)).ToEqual("2520")
		})
		t.Run("3", func(t tst.Test) {
			f := logftxt.DurationAsSeconds(logftxt.Precision(3))
			format := func(duration time.Duration) string {
				return string(f(nil, duration))
			}

			t.Expect(format(time.Second)).ToEqual("1.000")
			t.Expect(format(time.Millisecond)).ToEqual("0.001")
			t.Expect(format(time.Microsecond)).ToEqual("0.000")
			t.Expect(format(time.Nanosecond)).ToEqual("0.000")
			t.Expect(format(42 * time.Minute)).ToEqual("2520.000")
		})
	})

	t.Run("AsHMS", func(t tst.Test) {
		t.Run("Auto", func(t tst.Test) {
			f := logftxt.DurationAsHMS()
			format := func(duration time.Duration) string {
				return string(f(nil, duration))
			}

			t.Expect(format(time.Second)).ToEqual("00:00:01")
			t.Expect(format(time.Millisecond)).ToEqual("00:00:00.001")
			t.Expect(format(time.Microsecond)).ToEqual("00:00:00.000001")
			t.Expect(format(time.Nanosecond)).ToEqual("00:00:00.000000001")
			t.Expect(format(42 * time.Minute)).ToEqual("00:42:00")
			t.Expect(format(24*time.Hour + 42*time.Minute)).ToEqual("24:42:00")
			t.Expect(format(-42 * time.Minute)).ToEqual("-00:42:00")
		})
		t.Run("3", func(t tst.Test) {
			f := logftxt.DurationAsHMS(logftxt.Precision(3))
			format := func(duration time.Duration) string {
				return string(f(nil, duration))
			}

			t.Expect(format(time.Second)).ToEqual("00:00:01.000")
			t.Expect(format(time.Millisecond)).ToEqual("00:00:00.001")
			t.Expect(format(time.Microsecond)).ToEqual("00:00:00.000")
			t.Expect(format(time.Nanosecond)).ToEqual("00:00:00.000")
			t.Expect(format(42 * time.Minute)).ToEqual("00:42:00.000")
			t.Expect(format(-time.Millisecond)).ToEqual("-00:00:00.001")
		})

		t.Run("0", func(t tst.Test) {
			f := logftxt.DurationAsHMS(logftxt.Precision(0))
			format := func(duration time.Duration) string {
				return string(f(nil, duration))
			}

			t.Expect(format(time.Second)).ToEqual("00:00:01")
			t.Expect(format(time.Millisecond)).ToEqual("00:00:00")
			t.Expect(format(time.Microsecond)).ToEqual("00:00:00")
			t.Expect(format(time.Nanosecond)).ToEqual("00:00:00")
			t.Expect(format(42 * time.Minute)).ToEqual("00:42:00")
			t.Expect(format(-time.Millisecond)).ToEqual("-00:00:00")
		})
	})
}
