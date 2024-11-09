package logftxt_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ssgreg/logf"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt"
	"github.com/pamburus/logftxt/internal/pkg/pathx"
)

func TestEncoder(tt *testing.T) {
	t := tst.New(tt)

	envColor := func(value bool) logftxt.Environment {
		return func(name string) (string, bool) {
			if !value && name == "NO_COLOR" {
				return "1", true
			}

			return "", false
		}
	}

	config, err := logftxt.LoadConfig("./encoder_test.config.yml")
	t.Expect(err).ToNot(tst.HaveOccurred())

	theme := logftxt.NewThemeRef(pathx.ExplicitlyRelative("encoder_test.theme.yml"))
	caller := logf.EntryCaller{
		File:      "test.go",
		Line:      42,
		Specified: true,
	}

	t.Run("Level", func(t tst.Test) {
		enc := logftxt.NewEncoder(config, logftxt.ColorNever, theme)

		t.Run("BelowDebug", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelDebug + 1,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |DBG| msg\n")
		})

		t.Run("Debug", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelDebug,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |DBG| msg\n")
		})

		t.Run("Info", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelInfo,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |INF| msg\n")
		})

		t.Run("Warn", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelWarn,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |WRN| msg\n")
		})

		t.Run("Error", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelError,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |ERR| msg\n")
		})

		t.Run("AboveError", func(t tst.Test) {
			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, logf.Entry{
				Time:  time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				Level: logf.LevelError - 1,
				Text:  "msg",
			})).ToSucceed()
			t.Expect(buf.String()).ToEqual("Jan  2 03:04:05.000 |ERR| msg\n")
		})
	})

	t.Run("OccasionalComposite", func(t tst.Test) {
		testEntry := logf.Entry{
			LoggerName: "ml",
			DerivedFields: []logf.Field{
				logf.String("dsf1", "sv1"),
				logf.Int("dif1", 420),
				logf.Ints("dia", []int{420, 430, 440}),
				logf.Ints("dia0", []int{}),
				logf.String("dsf2", "sv2"),
				logf.String("dsf0", ""),
				logf.Object("o1", asObject(logf.String("a", "b"), logf.String("c", "d"))),
				logf.Int("dif1", 840),
				logf.Strings("dsa", []string{"abc", "def", "ghi"}),
			},
			Fields: []logf.Field{
				logf.String("sf", "sv"),
				logf.Int("if", 42),
				logf.Ints("array", []int{42, 43, 44}),
				logf.Object("o0", asObject()),
			},
			Level:  logf.LevelDebug,
			Time:   time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
			Text:   "The quick brown fox jumps over a lazy dog",
			Caller: caller,
		}

		t.Run("Flatten", func(t tst.Test) {
			enc := logftxt.NewEncoder(config, logftxt.ColorAlways, theme, logftxt.FlattenObjects(true))

			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, testEntry)).ToSucceed()

			t.Expect(buf.String()).ToEqual(
				"\x1b[2mJan  2 03:04:05.000\x1b[0m |\x1b[35mDBG\x1b[0m| \x1b[2mml:\x1b[0m \x1b[1mThe quick brown fox jumps over a lazy dog\x1b[0m \x1b[32mdsf1\x1b[0m\x1b[2m=\x1b[0msv1 \x1b[32mdif1\x1b[0m\x1b[2m=\x1b[0m\x1b[94m420\x1b[0m \x1b[32mdia\x1b[0m\x1b[2m=\x1b[0m[ \x1b[94m420\x1b[0m, \x1b[94m430\x1b[0m, \x1b[94m440\x1b[0m ] \x1b[32mdia0\x1b[0m\x1b[2m=\x1b[0m[] \x1b[32mdsf2\x1b[0m\x1b[2m=\x1b[0msv2 \x1b[32mdsf0\x1b[0m\x1b[2m=\x1b[0m\"\" \x1b[32mo1.a\x1b[0m\x1b[2m=\x1b[0mb \x1b[32mo1.c\x1b[0m\x1b[2m=\x1b[0md \x1b[32mdif1\x1b[0m\x1b[2m=\x1b[0m\x1b[94m840\x1b[0m \x1b[32mdsa\x1b[0m\x1b[2m=\x1b[0m[ abc, def, ghi ] \x1b[32msf\x1b[0m\x1b[2m=\x1b[0msv \x1b[32mif\x1b[0m\x1b[2m=\x1b[0m\x1b[94m42\x1b[0m \x1b[32marray\x1b[0m\x1b[2m=\x1b[0m[ \x1b[94m42\x1b[0m, \x1b[94m43\x1b[0m, \x1b[94m44\x1b[0m ] \x1b[90;3m@ test.go:42\x1b[0m\n",
			)
		})

		t.Run("NoFlatten", func(t tst.Test) {
			enc := logftxt.NewEncoder(config, logftxt.ColorAlways, theme, logftxt.FlattenObjects(false))

			buf := logf.NewBuffer()
			t.Expect(enc.Encode(buf, testEntry)).ToSucceed()

			t.Expect(buf.String()).ToEqual(
				"\x1b[2mJan  2 03:04:05.000\x1b[0m |\x1b[35mDBG\x1b[0m| \x1b[2mml:\x1b[0m \x1b[1mThe quick brown fox jumps over a lazy dog\x1b[0m \x1b[32mdsf1\x1b[0m\x1b[2m=\x1b[0msv1 \x1b[32mdif1\x1b[0m\x1b[2m=\x1b[0m\x1b[94m420\x1b[0m \x1b[32mdia\x1b[0m\x1b[2m=\x1b[0m[ \x1b[94m420\x1b[0m, \x1b[94m430\x1b[0m, \x1b[94m440\x1b[0m ] \x1b[32mdia0\x1b[0m\x1b[2m=\x1b[0m[] \x1b[32mdsf2\x1b[0m\x1b[2m=\x1b[0msv2 \x1b[32mdsf0\x1b[0m\x1b[2m=\x1b[0m\"\" \x1b[32mo1\x1b[0m\x1b[2m=\x1b[0m{ \x1b[32ma\x1b[0m\x1b[2m=\x1b[0mb, \x1b[32mc\x1b[0m\x1b[2m=\x1b[0md } \x1b[32mdif1\x1b[0m\x1b[2m=\x1b[0m\x1b[94m840\x1b[0m \x1b[32mdsa\x1b[0m\x1b[2m=\x1b[0m[ abc, def, ghi ] \x1b[32msf\x1b[0m\x1b[2m=\x1b[0msv \x1b[32mif\x1b[0m\x1b[2m=\x1b[0m\x1b[94m42\x1b[0m \x1b[32marray\x1b[0m\x1b[2m=\x1b[0m[ \x1b[94m42\x1b[0m, \x1b[94m43\x1b[0m, \x1b[94m44\x1b[0m ] \x1b[32mo0\x1b[0m\x1b[2m=\x1b[0m{} \x1b[90;3m@ test.go:42\x1b[0m\n",
			)
		})
	})

	someTime := time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)

	t.Run("Field", func(t tst.Test) {
		test := func(skip int, field logf.Field, key string, value string) func(t tst.Test) {
			tag := tst.CallerLine(skip + 1)

			return func(t tst.Test) {
				t.AddLineTags(tag)
				buf := logf.NewBuffer()
				appender := fixedTimestampAppender{logftxt.NewAppender(buf, config, theme, envColor(false), logftxt.FlattenObjects(false)), someTime}
				writer := logf.NewUnbufferedEntryWriter(appender)
				logger := logf.NewLogger(logf.LevelDebug, writer).WithName("me")
				logger.Info("msg", field)
				t.Expect(appender.Flush()).ToSucceed()
				t.Expect(buf.String()).ToEqual(
					fmt.Sprintf("Jan  2 03:04:05.000 |INF| me: msg %s=%s\n", key, value),
				)
			}
		}

		array := func(values ...string) string {
			return "[ " + strings.Join(values, ", ") + " ]"
		}

		object := func(pairs ...string) string {
			n := len(pairs) / 2
			if n == 0 {
				return "{}"
			}

			fields := make([]string, 0, n)
			for i := 0; i < len(pairs); i += 2 {
				fields = append(fields, fmt.Sprintf("%s=%s", pairs[i], pairs[i+1]))
			}

			return "{ " + strings.Join(fields, ", ") + " }"
		}

		t.Run("Int", test(0, logf.Int("a", 42), "a", "42"))
		t.Run("Int8", test(0, logf.Int8("a", 42), "a", "42"))
		t.Run("Int16", test(0, logf.Int16("a", 42), "a", "42"))
		t.Run("Int32", test(0, logf.Int32("a", 42), "a", "42"))
		t.Run("Int64", test(0, logf.Int64("a", 42), "a", "42"))
		t.Run("Uint", test(0, logf.Uint("a", 42), "a", "42"))
		t.Run("Uint8", test(0, logf.Uint8("a", 42), "a", "42"))
		t.Run("Uint16", test(0, logf.Uint16("a", 42), "a", "42"))
		t.Run("Uint32", test(0, logf.Uint32("a", 42), "a", "42"))
		t.Run("Uint64", test(0, logf.Uint64("a", 42), "a", "42"))
		t.Run("Bool", test(0, logf.Bool("a", true), "a", "true"))
		t.Run("Float32", test(0, logf.Float32("a", 42.43), "a", "42.43"))
		t.Run("Float64", test(0, logf.Float64("a", 42.43), "a", "42.43"))
		t.Run("String", func(t tst.Test) {
			t.Run("Simple", test(0, logf.String("a", "oue"), "a", "oue"))
			t.Run("Empty", test(0, logf.String("a", ""), "a", `""`))
			t.Run("WithEqualSign", test(0, logf.String("a", "a=b"), "a", `"a=b"`))
			t.Run("WithSpace", test(0, logf.String("a", "a b"), "a", `"a b"`))
			t.Run("WithNewLine", test(0, logf.String("a", "a\nb"), "a", `"a\nb"`))
			t.Run("WithDot", test(0, logf.String("a", "a.b"), "a", "a.b"))
			t.Run("LikeInt", test(0, logf.String("a", "42"), "a", `"42"`))
			t.Run("LikeFloat", test(0, logf.String("a", "42.1"), "a", `"42.1"`))
			t.Run("LikeIPAddress", test(0, logf.String("a", "192.168.0.1"), "a", "192.168.0.1"))
			t.Run("Null", test(0, logf.String("a", "null"), "a", `"null"`))
			t.Run("Special", test(0, logf.String("a", "-\x00-\n-\r-\t-\\-\xf3-\"\xff\"-ё"), "a", `"-\u0000-\n-\r-\t-\\-\ufffd-\"\ufffd\"-ё"`))
		})
		t.Run("Duration", test(0, logf.Duration("a", time.Second), "a", "00:00:01"))
		t.Run("Error", test(0, logf.Error(errors.New("oops")), "error", "{{ oops }}"))
		t.Run("NilError", test(0, logf.Error(nil), "error", "{{ <nil> }}"))
		t.Run("Time", test(0, logf.Time("t", someTime), "t", "[[Jan  2 03:04:05.000]]"))
		t.Run("Ints", test(0, logf.Ints("a", []int{42}), "a", array("42")))
		t.Run("Ints8", test(0, logf.Ints8("a", []int8{42}), "a", array("42")))
		t.Run("Ints16", test(0, logf.Ints16("a", []int16{42}), "a", array("42")))
		t.Run("Ints32", test(0, logf.Ints32("a", []int32{42}), "a", array("42")))
		t.Run("Ints64", test(0, logf.Ints64("a", []int64{42}), "a", array("42")))
		t.Run("Uints", test(0, logf.Uints("a", []uint{42}), "a", array("42")))
		t.Run("Uints8", test(0, logf.Uints8("a", []uint8{42}), "a", array("42")))
		t.Run("Uints16", test(0, logf.Uints16("a", []uint16{42}), "a", array("42")))
		t.Run("Uints32", test(0, logf.Uints32("a", []uint32{42}), "a", array("42")))
		t.Run("Uints64", test(0, logf.Uints64("a", []uint64{42}), "a", array("42")))
		t.Run("Bools", test(0, logf.Bools("a", []bool{true, false}), "a", array("true", "false")))
		t.Run("Floats32", test(0, logf.Floats32("a", []float32{42.43, 42.42}), "a", array("42.43", "42.42")))
		t.Run("Floats64", test(0, logf.Floats64("a", []float64{42.43, 42.42}), "a", array("42.43", "42.42")))
		t.Run("Durations", test(0, logf.Durations("a", []time.Duration{time.Second, time.Minute}), "a", array("00:00:01", "00:01:00")))
		t.Run("Strings", test(0, logf.Strings("a", []string{"bcd", "efg"}), "a", array("bcd", "efg")))
		t.Run("StringsEmpty", test(0, logf.Strings("a", []string{}), "a", "[]"))
		t.Run("Bytes", test(0, logf.Bytes("a", []byte{1, 2, 3}), "a", "AQID"))

		t.Run("Any", func(t tst.Test) {
			parentTest := test
			test := func(skip int, key string, actual any, expected string) func(t tst.Test) {
				return parentTest(skip+1, logf.Array(key, newMockArray(logf.TypeEncoder.EncodeTypeAny, actual)), key, array(expected))
			}

			t.Run("Nil", test(0, "a", nil, "null"))
			t.Run("Int", test(0, "a", int(42), "42"))
			t.Run("Int8", test(0, "a", int8(42), "42"))
			t.Run("Int16", test(0, "a", int16(42), "42"))
			t.Run("Int32", test(0, "a", int32(42), "42"))
			t.Run("Int64", test(0, "a", int64(42), "42"))
			t.Run("Uint", test(0, "a", uint(42), "42"))
			t.Run("Uint8", test(0, "a", uint8(42), "42"))
			t.Run("Uint16", test(0, "a", uint16(42), "42"))
			t.Run("Uint32", test(0, "a", uint32(42), "42"))
			t.Run("Uint64", test(0, "a", uint64(42), "42"))
			t.Run("Bool", test(0, "a", bool(true), "true"))
			t.Run("Float32", test(0, "a", float32(42.43), "42.43"))
			t.Run("Float64", test(0, "a", float64(42.43), "42.43"))
			t.Run("String", test(0, "a", "oue", "oue"))
			t.Run("Duration", test(0, "a", time.Second, "00:00:01"))
			t.Run("Error", test(0, "e", errors.New("oops"), "{{ oops }}"))
			t.Run("Time", test(0, "t", someTime, "[[Jan  2 03:04:05.000]]"))
			t.Run("Ints", test(0, "a", []int{42}, array("42")))
			t.Run("Ints8", test(0, "a", []int8{42}, array("42")))
			t.Run("Ints16", test(0, "a", []int16{42}, array("42")))
			t.Run("Ints32", test(0, "a", []int32{42}, array("42")))
			t.Run("Ints64", test(0, "a", []int64{42}, array("42")))
			t.Run("Uints", test(0, "a", []uint{42}, array("42")))
			t.Run("Uints8", test(0, "a", []uint8{42}, "Kg=="))
			t.Run("Uints16", test(0, "a", []uint16{42}, array("42")))
			t.Run("Uints32", test(0, "a", []uint32{42}, array("42")))
			t.Run("Uints64", test(0, "a", []uint64{42}, array("42")))
			t.Run("Bools", test(0, "a", []bool{true, false}, array("true", "false")))
			t.Run("Floats32", test(0, "a", []float32{42.43, 42.42}, array("42.43", "42.42")))
			t.Run("Floats64", test(0, "a", []float64{42.43, 42.42}, array("42.43", "42.42")))
			t.Run("Durations", test(0, "a", []time.Duration{time.Second, time.Minute}, array("00:00:01", "00:01:00")))
			t.Run("Strings", test(0, "a", []string{"bcd", "efg"}, array("bcd", "efg")))
			t.Run("Bytes", test(0, "a", []byte{1, 2, 3}, "AQID"))
			t.Run("Array", test(0, "a", newMockArray(logf.TypeEncoder.EncodeTypeBool, false, true), array("false", "true")))
			t.Run("Object", test(0, "a", newMockObject(logf.Bool("b", true)), object("b", "true")))
			t.Run("Stringer", test(0, "a", mockStringer("oue"), "oue"))
			t.Run("AnySlice", test(0, "a", []newTypeString{"2", "42"}, array(`"2"`, `"42"`)))
			t.Run("AnyArray", test(0, "a", [2]newTypeString{"2", "42"}, array(`"2"`, `"42"`)))

			t.Run("NewType", func(t tst.Test) {
				t.Run("Bool", test(0, "a", newTypeBool(true), "true"))
				t.Run("Int", test(0, "a", newTypeInt(42), "42"))
				t.Run("Int8", test(0, "a", newTypeInt8(42), "42"))
				t.Run("Int16", test(0, "a", newTypeInt16(42), "42"))
				t.Run("Int32", test(0, "a", newTypeInt32(42), "42"))
				t.Run("Int64", test(0, "a", newTypeInt64(42), "42"))
				t.Run("Uint", test(0, "a", newTypeUint(42), "42"))
				t.Run("Uint8", test(0, "a", newTypeUint8(42), "42"))
				t.Run("Uint16", test(0, "a", newTypeUint16(42), "42"))
				t.Run("Uint32", test(0, "a", newTypeUint32(42), "42"))
				t.Run("Uint64", test(0, "a", newTypeUint64(42), "42"))
				t.Run("Float32", test(0, "a", newTypeFloat32(42.2), "42.2"))
				t.Run("Float64", test(0, "a", newTypeFloat64(42.3), "42.3"))
				t.Run("String", test(0, "a", newTypeString("abc"), "abc"))
			})
		})

		t.Run("Array", func(t tst.Test) {
			type te = logf.TypeEncoder
			parentTest := test
			test := func(skip int, key string, actual logf.ArrayEncoder, expected ...string) func(t tst.Test) {
				return parentTest(skip+1, logf.Array(key, actual), key, array(expected...))
			}

			t.Run("Bool", test(0, "a", newMockArray(te.EncodeTypeBool, true, false), "true", "false"))
			t.Run("Int8", test(0, "a", newMockArray(te.EncodeTypeInt8, 1, 2, 3), "1", "2", "3"))
			t.Run("Int16", test(0, "a", newMockArray(te.EncodeTypeInt16, 1, 2, 3), "1", "2", "3"))
			t.Run("Int32", test(0, "a", newMockArray(te.EncodeTypeInt32, 1, 2, 3), "1", "2", "3"))
			t.Run("Int64", test(0, "a", newMockArray(te.EncodeTypeInt64, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint8", test(0, "a", newMockArray(te.EncodeTypeUint8, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint16", test(0, "a", newMockArray(te.EncodeTypeUint16, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint32", test(0, "a", newMockArray(te.EncodeTypeUint32, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint64", test(0, "a", newMockArray(te.EncodeTypeUint64, 1, 2, 3), "1", "2", "3"))
			t.Run("Float32", test(0, "a", newMockArray(te.EncodeTypeFloat32, 1.2, 2.3), "1.2", "2.3"))
			t.Run("Float64", test(0, "a", newMockArray(te.EncodeTypeFloat64, 1.2, 2.3), "1.2", "2.3"))
			t.Run("Duration", test(0, "a", newMockArray(te.EncodeTypeDuration, time.Hour, time.Minute), "01:00:00", "00:01:00"))
			t.Run("Time", test(0, "a", newMockArray(te.EncodeTypeTime, someTime), "[[Jan  2 03:04:05.000]]"))
			t.Run("Bytes", test(0, "a", newMockArray(te.EncodeTypeBytes, []byte{1, 2, 3}), "AQID"))
			t.Run("Ints8", test(0, "a", newMockArray(te.EncodeTypeInts8, []int8{42}), array("42")))
			t.Run("Ints16", test(0, "a", newMockArray(te.EncodeTypeInts16, []int16{42}), array("42")))
			t.Run("Ints32", test(0, "a", newMockArray(te.EncodeTypeInts32, []int32{42}), array("42")))
			t.Run("Ints64", test(0, "a", newMockArray(te.EncodeTypeInts64, []int64{42}), array("42")))
			t.Run("Uints8", test(0, "a", newMockArray(te.EncodeTypeUints8, []uint8{42}), array("42")))
			t.Run("Uints16", test(0, "a", newMockArray(te.EncodeTypeUints16, []uint16{42}), array("42")))
			t.Run("Uints32", test(0, "a", newMockArray(te.EncodeTypeUints32, []uint32{42}), array("42")))
			t.Run("Uints64", test(0, "a", newMockArray(te.EncodeTypeUints64, []uint64{42}), array("42")))
			t.Run("Bools", test(0, "a", newMockArray(te.EncodeTypeBools, []bool{true, false}), array("true", "false")))
			t.Run("Floats32", test(0, "a", newMockArray(te.EncodeTypeFloats32, []float32{42.43, 42.42}), array("42.43", "42.42")))
			t.Run("Floats64", test(0, "a", newMockArray(te.EncodeTypeFloats64, []float64{42.43, 42.42}), array("42.43", "42.42")))
			t.Run("Durations", test(0, "a", newMockArray(te.EncodeTypeDurations, []time.Duration{time.Second, time.Minute}), array("00:00:01", "00:01:00")))
			t.Run("Strings", test(0, "a", newMockArray(te.EncodeTypeStrings, []string{"bcd", "efg"}), array("bcd", "efg")))
			t.Run("Array", test(0, "a", newMockArray(te.EncodeTypeArray, newMockArray(logf.TypeEncoder.EncodeTypeInt8, 42)), array("42")))
			t.Run("Object", test(0, "a", newMockArray(te.EncodeTypeObject, newMockObject(logf.Object("b", newMockObject(logf.Int("c", 42))))), object("b", object("c", "42"))))
		})

		t.Run("Object", func(t tst.Test) {
			parentTest := test
			test := func(skip int, key string, actual logf.ObjectEncoder, expected ...string) func(t tst.Test) {
				return parentTest(skip+1, logf.Object(key, actual), key, object(expected...))
			}

			t.Run("Empty", test(0, "a", newMockObject()))
			t.Run("Bool", test(0, "a", newMockObject(logf.Bool("b", true), logf.Bool("c", false)), "b", "true", "c", "false"))
			t.Run("Int8", test(0, "a", newMockObject(logf.Int8("b", 1)), "b", "1"))
			t.Run("Int16", test(0, "a", newMockObject(logf.Int16("b", 2)), "b", "2"))
			t.Run("Int32", test(0, "a", newMockObject(logf.Int32("b", 3)), "b", "3"))
			t.Run("Int64", test(0, "a", newMockObject(logf.Int64("b", 42)), "b", "42"))
			t.Run("Uint8", test(0, "a", newMockObject(logf.Uint8("b", 43)), "b", "43"))
			t.Run("Uint16", test(0, "a", newMockObject(logf.Uint16("b", 3)), "b", "3"))
			t.Run("Uint32", test(0, "a", newMockObject(logf.Uint32("b", 3)), "b", "3"))
			t.Run("Uint64", test(0, "a", newMockObject(logf.Uint64("b", 3)), "b", "3"))
			t.Run("Float32", test(0, "a", newMockObject(logf.Float32("b", 2.3)), "b", "2.3"))
			t.Run("Float64", test(0, "a", newMockObject(logf.Float64("b", 2.3)), "b", "2.3"))
			t.Run("Duration", test(0, "a", newMockObject(logf.Duration("b", time.Hour)), "b", "01:00:00"))
			t.Run("Time", test(0, "a", newMockObject(logf.Time("b", someTime)), "b", "[[Jan  2 03:04:05.000]]"))
			t.Run("Bytes", test(0, "a", newMockObject(logf.Bytes("b", []byte{1, 2, 3})), "b", "AQID"))
			t.Run("String", test(0, "a", newMockObject(logf.String("b", "c")), "b", "c"))
			t.Run("Error", test(0, "a", newMockObject(logf.NamedError("b", errors.New("c"))), "b", "{{ c }}"))
			t.Run("Ints8", test(0, "a", newMockObject(logf.Ints8("b", []int8{42})), "b", array("42")))
			t.Run("Ints16", test(0, "a", newMockObject(logf.Ints16("b", []int16{42})), "b", array("42")))
			t.Run("Ints32", test(0, "a", newMockObject(logf.Ints32("b", []int32{42})), "b", array("42")))
			t.Run("Ints64", test(0, "a", newMockObject(logf.Ints64("b", []int64{42})), "b", array("42")))
			t.Run("Uints8", test(0, "a", newMockObject(logf.Uints8("b", []uint8{42})), "b", array("42")))
			t.Run("Uints16", test(0, "a", newMockObject(logf.Uints16("b", []uint16{42})), "b", array("42")))
			t.Run("Uints32", test(0, "a", newMockObject(logf.Uints32("b", []uint32{42})), "b", array("42")))
			t.Run("Uints64", test(0, "a", newMockObject(logf.Uints64("b", []uint64{42})), "b", array("42")))
			t.Run("Bools", test(0, "a", newMockObject(logf.Bools("b", []bool{true, false})), "b", array("true", "false")))
			t.Run("Floats32", test(0, "a", newMockObject(logf.Floats32("b", []float32{42.43, 42.42})), "b", array("42.43", "42.42")))
			t.Run("Floats64", test(0, "a", newMockObject(logf.Floats64("b", []float64{42.43, 42.42})), "b", array("42.43", "42.42")))
			t.Run("Durations", test(0, "a", newMockObject(logf.Durations("b", []time.Duration{time.Second, time.Minute})), "b", array("00:00:01", "00:01:00")))
			t.Run("Strings", test(0, "a", newMockObject(logf.Strings("b", []string{"bcd", "efg"})), "b", array("bcd", "efg")))
			t.Run("Array", test(0, "a", newMockObject(logf.Array("b", newMockArray(logf.TypeEncoder.EncodeTypeInt8, 42))), "b", array("42")))
			t.Run("Object", test(0, "a", newMockObject(logf.Object("b", newMockObject(logf.Int("c", 42)))), "b", object("c", "42")))
			t.Run("Any", test(0, "a", newMockObject(logf.Any("b", anyStruct{42})), "b", "{42}"))
		})
	})

	t.Run("FieldFlatten", func(t tst.Test) {
		test := func(skip int, field logf.Field, key string, value string) func(t tst.Test) {
			tag := tst.CallerLine(skip + 1)

			return func(t tst.Test) {
				t.AddLineTags(tag)
				buf := logf.NewBuffer()
				appender := fixedTimestampAppender{logftxt.NewAppender(buf, config, theme, envColor(false), logftxt.FlattenObjects(true)), someTime}
				writer := logf.NewUnbufferedEntryWriter(appender)
				logger := logf.NewLogger(logf.LevelDebug, writer).WithName("me")
				logger.Info("msg", field)
				field := ""
				if value != "" {
					if key != "" {
						field = fmt.Sprintf(" %s=%s", key, value)
					} else {
						field = " " + value
					}
				}
				t.Expect(appender.Flush()).ToSucceed()
				t.Expect(buf.String()).ToEqual(
					fmt.Sprintf("Jan  2 03:04:05.000 |INF| me: msg%s\n", field),
				)
			}
		}

		array := func(values ...string) string {
			return "[ " + strings.Join(values, ", ") + " ]"
		}

		outerObject := func(pairs ...string) string {
			n := len(pairs) / 2
			if n == 0 {
				return ""
			}

			fields := make([]string, 0, n)
			for i := 0; i < len(pairs); i += 2 {
				fields = append(fields, fmt.Sprintf("%s=%s", pairs[i], pairs[i+1]))
			}

			return strings.Join(fields, " ")
		}

		innerObject := func(pairs ...string) string {
			n := len(pairs) / 2
			if n == 0 {
				return "{}"
			}

			return "{ " + outerObject(pairs...) + " }"
		}

		t.Run("Int", test(0, logf.Int("a", 42), "a", "42"))
		t.Run("String", test(0, logf.String("a", "oue"), "a", "oue"))
		t.Run("Strings", test(0, logf.Strings("a", []string{"bcd", "efg"}), "a", array("bcd", "efg")))

		t.Run("Any", func(t tst.Test) {
			parentTest := test
			test := func(skip int, key string, actual any, expected string) func(t tst.Test) {
				return parentTest(skip+1, logf.Array(key, newMockArray(logf.TypeEncoder.EncodeTypeAny, actual)), key, array(expected))
			}

			t.Run("Nil", test(0, "a", nil, "null"))
			t.Run("Int", test(0, "a", int(42), "42"))
			t.Run("Bool", test(0, "a", bool(true), "true"))
			t.Run("Array", test(0, "a", newMockArray(logf.TypeEncoder.EncodeTypeBool, false, true), array("false", "true")))
			t.Run("Object", test(0, "a", newMockObject(logf.Bool("b", true)), innerObject("b", "true")))
			t.Run("Stringer", test(0, "a", mockStringer("oue"), "oue"))
			t.Run("AnySlice", test(0, "a", []newTypeString{"2", "42"}, array(`"2"`, `"42"`)))
			t.Run("AnyArray", test(0, "a", [2]newTypeString{"2", "42"}, array(`"2"`, `"42"`)))

			t.Run("NewType", func(t tst.Test) {
				t.Run("Int", test(0, "a", newTypeInt(42), "42"))
				t.Run("String", test(0, "a", newTypeString("abc"), "abc"))
			})
		})

		t.Run("Array", func(t tst.Test) {
			type te = logf.TypeEncoder
			parentTest := test
			test := func(skip int, key string, actual logf.ArrayEncoder, expected ...string) func(t tst.Test) {
				return parentTest(skip+1, logf.Array(key, actual), key, array(expected...))
			}

			t.Run("Bool", test(0, "a", newMockArray(te.EncodeTypeBool, true, false), "true", "false"))
			t.Run("Int8", test(0, "a", newMockArray(te.EncodeTypeInt8, 1, 2, 3), "1", "2", "3"))
			t.Run("Int16", test(0, "a", newMockArray(te.EncodeTypeInt16, 1, 2, 3), "1", "2", "3"))
			t.Run("Int32", test(0, "a", newMockArray(te.EncodeTypeInt32, 1, 2, 3), "1", "2", "3"))
			t.Run("Int64", test(0, "a", newMockArray(te.EncodeTypeInt64, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint8", test(0, "a", newMockArray(te.EncodeTypeUint8, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint16", test(0, "a", newMockArray(te.EncodeTypeUint16, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint32", test(0, "a", newMockArray(te.EncodeTypeUint32, 1, 2, 3), "1", "2", "3"))
			t.Run("Uint64", test(0, "a", newMockArray(te.EncodeTypeUint64, 1, 2, 3), "1", "2", "3"))
			t.Run("Float32", test(0, "a", newMockArray(te.EncodeTypeFloat32, 1.2, 2.3), "1.2", "2.3"))
			t.Run("Float64", test(0, "a", newMockArray(te.EncodeTypeFloat64, 1.2, 2.3), "1.2", "2.3"))
			t.Run("Duration", test(0, "a", newMockArray(te.EncodeTypeDuration, time.Hour, time.Minute), "01:00:00", "00:01:00"))
			t.Run("Time", test(0, "a", newMockArray(te.EncodeTypeTime, someTime), "[[Jan  2 03:04:05.000]]"))
			t.Run("Bytes", test(0, "a", newMockArray(te.EncodeTypeBytes, []byte{1, 2, 3}), "AQID"))
			t.Run("Ints8", test(0, "a", newMockArray(te.EncodeTypeInts8, []int8{42}), array("42")))
			t.Run("Ints16", test(0, "a", newMockArray(te.EncodeTypeInts16, []int16{42}), array("42")))
			t.Run("Ints32", test(0, "a", newMockArray(te.EncodeTypeInts32, []int32{42}), array("42")))
			t.Run("Ints64", test(0, "a", newMockArray(te.EncodeTypeInts64, []int64{42}), array("42")))
			t.Run("Uints8", test(0, "a", newMockArray(te.EncodeTypeUints8, []uint8{42}), array("42")))
			t.Run("Uints16", test(0, "a", newMockArray(te.EncodeTypeUints16, []uint16{42}), array("42")))
			t.Run("Uints32", test(0, "a", newMockArray(te.EncodeTypeUints32, []uint32{42}), array("42")))
			t.Run("Uints64", test(0, "a", newMockArray(te.EncodeTypeUints64, []uint64{42}), array("42")))
			t.Run("Bools", test(0, "a", newMockArray(te.EncodeTypeBools, []bool{true, false}), array("true", "false")))
			t.Run("Floats32", test(0, "a", newMockArray(te.EncodeTypeFloats32, []float32{42.43, 42.42}), array("42.43", "42.42")))
			t.Run("Floats64", test(0, "a", newMockArray(te.EncodeTypeFloats64, []float64{42.43, 42.42}), array("42.43", "42.42")))
			t.Run("Durations", test(0, "a", newMockArray(te.EncodeTypeDurations, []time.Duration{time.Second, time.Minute}), array("00:00:01", "00:01:00")))
			t.Run("Strings", test(0, "a", newMockArray(te.EncodeTypeStrings, []string{"bcd", "efg"}), array("bcd", "efg")))
			t.Run("Array", test(0, "a", newMockArray(te.EncodeTypeArray, newMockArray(logf.TypeEncoder.EncodeTypeInt8, 42)), array("42")))
			t.Run("Object", test(0, "a", newMockArray(te.EncodeTypeObject, newMockObject(logf.Object("b", newMockObject(logf.Int("c", 42))))), innerObject("b.c", "42")))
		})

		t.Run("Object", func(t tst.Test) {
			parentTest := test
			test := func(skip int, key string, actual logf.ObjectEncoder, expected ...string) func(t tst.Test) {
				return parentTest(skip+1, logf.Object(key, actual), "", outerObject(expected...))
			}

			t.Run("Empty", test(0, "a", newMockObject()))
			t.Run("Bool", test(0, "a", newMockObject(logf.Bool("b", true), logf.Bool("c", false)), "a.b", "true", "a.c", "false"))
			t.Run("Strings", test(0, "a", newMockObject(logf.Strings("b", []string{"bcd", "efg"})), "a.b", array("bcd", "efg")))
			t.Run("Array", test(0, "a", newMockObject(logf.Array("b", newMockArray(logf.TypeEncoder.EncodeTypeInt8, 42))), "a.b", array("42")))
			t.Run("Object", test(0, "a", newMockObject(logf.Object("b", newMockObject(logf.Int("c", 42), logf.Object("d", newMockObject())))), "a.b.c", "42"))
			t.Run("Any", test(0, "a", newMockObject(logf.Any("b", anyStruct{42})), "a.b", "{42}"))
		})
	})

	t.Run("ConfigProvideError", func(t tst.Test) {
		enc := logftxt.NewEncoder(
			config,
			theme,
			envColor(false),
			logftxt.FlattenObjects(true),
			logftxt.ConfigProvideFunc(func(logftxt.Domain) (*logftxt.Config, error) {
				return nil, errors.New("cperr")
			}),
		)

		buf := logf.NewBuffer()

		t.Expect(enc.Encode(buf, logf.Entry{
			Text: "msg",
		})).ToSucceed()

		t.Expect(buf.String()).ToEqual("Dec 31 23:59:59.999 |WRN| logftxt: failed to load configuration file so using previous defaults error={{ cperr }}\nJan  1 00:00:00.000 |ERR| msg\n")
	})

	t.Run("ThemeProvideError", func(t tst.Test) {
		theme, err := theme.Load()
		t.Expect(err).ToNot(tst.HaveOccurred())

		enc := logftxt.NewEncoder(
			config,
			theme,
			envColor(false),
			logftxt.ThemeProvideFunc(func(logftxt.Domain) (*logftxt.Theme, error) {
				return nil, errors.New("tperr")
			}),
		)

		buf := logf.NewBuffer()

		t.Expect(enc.Encode(buf, logf.Entry{
			Text: "msg",
		})).ToSucceed()

		t.Expect(buf.String()).ToEqual("Dec 31 23:59:59.999 |WRN| logftxt: failed to setup preferred theme so using previous defaults error={{ tperr }}\nJan  1 00:00:00.000 |ERR| msg\n")
	})

	t.Run("NoThemeSetting", func(t tst.Test) {
		enc := logftxt.NewEncoder(envColor(false), &logftxt.Config{})
		buf := logf.NewBuffer()
		t.Expect(enc.Encode(buf, logf.Entry{
			Text: "msg",
		})).ToSucceed()

		t.Expect(buf.String()).ToEqual("[ERR] msg\n")
	})

	t.Run("Config", func(t tst.Test) {
		test := func(t tst.Test, cfg logftxt.Config, field logf.Field, expected string, mw ...func(*logf.Entry)) {
			t.Helper()
			enc := logftxt.NewEncoder(cfg, envColor(false), theme)
			buf := logf.NewBuffer()
			entry := logf.Entry{
				Text:   "msg",
				Fields: []logf.Field{field},
			}
			for _, mw := range mw {
				mw(&entry)
			}
			t.Expect(enc.Encode(buf, entry)).ToSucceed()
			t.Expect(buf.String()).ToEqual(fmt.Sprintf("|ERR| msg %s\n", expected))
		}

		t.Run("Duration", func(t tst.Test) {
			t.Run("HMS", func(t tst.Test) {
				cfg := logftxt.Config{}
				cfg.Values.Duration.Format = logftxt.DurationFormatHMS
				test(t, cfg, logf.Duration("a", time.Minute), "a=00:01:00")
			})

			t.Run("Seconds", func(t tst.Test) {
				cfg := logftxt.Config{}
				cfg.Values.Duration.Format = logftxt.DurationFormatSeconds
				test(t, cfg, logf.Duration("a", time.Minute), "a=60")
			})
			t.Run("Dynamic", func(t tst.Test) {
				cfg := logftxt.Config{}
				cfg.Values.Duration.Format = logftxt.DurationFormatDynamic
				test(t, cfg, logf.Duration("a", time.Minute), "a=1m0s")
			})
		})

		t.Run("Caller", func(t tst.Test) {
			t.Run("Long", func(t tst.Test) {
				cfg := logftxt.Config{}
				cfg.Caller.Format = logftxt.CallerFormatLong
				test(t, cfg, logf.Int("a", 1), "a=1 @ /a/b/c.go:42", func(e *logf.Entry) {
					e.Caller.File = "/a/b/c.go"
					e.Caller.Line = 42
					e.Caller.Specified = true
				})
			})
		})

		t.Run("Error", func(t tst.Test) {
			t.Run("Long", func(t tst.Test) {
				cfg := logftxt.Config{}
				cfg.Values.Error.Format = logftxt.ErrorFormatLong
				test(t, cfg, logf.NamedError("e", mockError{}), "e={{ detailed error }}")
			})
		})
	})
}

// ---

type fixedTimestampAppender struct {
	logf.Appender
	ts time.Time
}

func (a fixedTimestampAppender) Append(e logf.Entry) error {
	e.Time = a.ts

	return a.Appender.Append(e)
}

// ---

func newMockArray[T any](f func(logf.TypeEncoder, T), values ...T) logf.ArrayEncoder {
	return mockArray[T]{f, values}
}

type mockArray[T any] struct {
	f      func(logf.TypeEncoder, T)
	values []T
}

func (a mockArray[T]) EncodeLogfArray(enc logf.TypeEncoder) error {
	for _, value := range a.values {
		a.f(enc, value)
	}

	return nil
}

// ---

func newMockObject(fields ...logf.Field) logf.ObjectEncoder {
	return mockObject{fields}
}

type mockObject struct {
	fields []logf.Field
}

func (a mockObject) EncodeLogfObject(enc logf.FieldEncoder) error {
	for _, field := range a.fields {
		field.Accept(enc)
	}

	return nil
}

// ---

type mockStringer string

func (s mockStringer) String() string {
	return string(s)
}

//

func asObject(fields ...logf.Field) logf.ObjectEncoder {
	return &fieldsAsObject{fields}
}

// ---

type fieldsAsObject struct {
	fields []logf.Field
}

func (f *fieldsAsObject) EncodeLogfObject(enc logf.FieldEncoder) error {
	for _, field := range f.fields {
		field.Accept(enc)
	}

	return nil
}

// ---

type anyStruct struct {
	a int
}

type newTypeBool bool
type newTypeInt int
type newTypeInt8 int8
type newTypeInt16 int16
type newTypeInt32 int32
type newTypeInt64 int64
type newTypeUint uint
type newTypeUint8 uint8
type newTypeUint16 uint16
type newTypeUint32 uint32
type newTypeUint64 uint64
type newTypeFloat32 float32
type newTypeFloat64 float64
type newTypeString string
