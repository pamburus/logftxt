package benchmarks

import (
	"testing"
	"time"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"

	"github.com/pamburus/logftxt"
)

func BenchmarkEncoder(b *testing.B) {
	buf := logf.NewBufferWithCapacity(1024 * 1024)

	b.Run("logftxt", func(b *testing.B) {
		b.Run("same-logger-id", func(b *testing.B) {
			encoder := logftxt.NewEncoder()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				buf.Reset()
				err := encoder.Encode(buf, testEntry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("new-logger-id", func(b *testing.B) {
			encoder := logftxt.NewEncoder()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				buf.Reset()
				testEntry.LoggerID++
				err := encoder.Encode(buf, testEntry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("logftext", func(b *testing.B) {
		b.Run("same-logger-id", func(b *testing.B) {
			encoder := logftext.NewEncoder.Default()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				buf.Reset()
				err := encoder.Encode(buf, testEntry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("new-logger-id", func(b *testing.B) {
			encoder := logftext.NewEncoder.Default()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i != b.N; i++ {
				buf.Reset()
				testEntry.LoggerID++
				err := encoder.Encode(buf, testEntry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// ---

var testEntry = logf.Entry{
	LoggerName: "some.test.logger",
	DerivedFields: []logf.Field{
		logf.String("derived-string-field-1", "string-value-1"),
		logf.Int("derived-int-field", 420),
		logf.Ints("derived-int-array", []int{420, 430, 440}),
		logf.String("derived-string-field-2", "string-value-2"),
		logf.Int("derived-int-field", 840),
		logf.Strings("derived-string-array", []string{"abc", "def", "ghi"}),
	},
	Fields: []logf.Field{
		logf.String("string-field", "string-value"),
		logf.Int("int-field", 42),
		logf.Ints("array", []int{42, 43, 44}),
	},
	Level:  logf.LevelDebug,
	Time:   time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
	Text:   "The quick brown fox jumps over a lazy dog",
	Caller: logf.NewEntryCaller(0),
}
