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

	testOne := func(b *testing.B, encoder logf.Encoder, hook func(*logf.Entry)) {
		entry := testEntryTemplate

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i != b.N; i++ {
			buf.Reset()
			hook(&entry)
			err := encoder.Encode(buf, entry)
			if err != nil {
				b.Fatal(err)
			}
		}
	}

	testSameAndNew := func(encoder logf.Encoder) func(b *testing.B) {
		return func(b *testing.B) {
			b.Run("SameLoggerID", func(b *testing.B) {
				testOne(b, encoder, func(entry *logf.Entry) {})
			})
			b.Run("NewLoggerID", func(b *testing.B) {
				testOne(b, encoder, func(entry *logf.Entry) {
					entry.LoggerID++
				})
			})
		}
	}

	b.Run("logftxt", func(b *testing.B) {
		options := []logftxt.EncoderOption{
			logftxt.CallerShort(),
			logftxt.NewThemeRef("@legacy"),
			logftxt.FlattenObjects(true),
		}

		b.Run("WithColor", testSameAndNew(logftxt.NewEncoder(append(options, logftxt.ColorAlways)...)))
		b.Run("WithoutColor", testSameAndNew(logftxt.NewEncoder(append(options, logftxt.ColorNever)...)))
	})

	b.Run("logftext", func(b *testing.B) {
		config := func(noColor bool) logftext.EncoderConfig {
			return logftext.EncoderConfig{
				NoColor: &noColor,
			}
		}

		b.Run("WithColor", testSameAndNew(logftext.NewEncoder(config(false))))
		b.Run("WithoutColor", testSameAndNew(logftext.NewEncoder(config(true))))
	})
}

var testEntryTemplate = logf.Entry{
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
