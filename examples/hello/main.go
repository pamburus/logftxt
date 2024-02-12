package main

import (
	"errors"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/ssgreg/logf"

	"github.com/pamburus/logftxt"
)

func main() {
	// Create ChannelWriter with text Encoder with default settings.
	writer, writerClose := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logftxt.NewAppender(os.Stdout, logftxt.FlattenObjects(true)),
	})
	defer writerClose()

	// Create Logger with ChannelWriter.
	logger := logf.NewLogger(logf.LevelDebug, writer).WithCaller().WithName("main")

	// Do some logging.
	logger.Info("runtime info", logf.Int("cpu-count", runtime.NumCPU()))
	logger.Info("test array", logf.Any("ss", []string{"cpu-count", "abc", "", "x"}))

	if info, ok := debug.ReadBuildInfo(); ok {
		logger.Debug("build info", logf.String("go-version", info.GoVersion), logf.String("path", info.Path))
		for _, setting := range info.Settings {
			logger.Debug("build setting", logf.String("key", setting.Key), logf.String("value", setting.Value))
		}

		logger.Warn("dependencies found", logf.Int("count", len(info.Deps)))

		for _, dep := range info.Deps {
			logger.Info("dependency",
				logf.Object("info", AsObject(logf.String("path", dep.Path), logf.String("version", dep.Version))),
				logf.Object("extra", AsObject(
					logf.Object("a", AsObject(
						logf.Object("g", AsObject(
							logf.String("b", "c"),
						)),
						logf.String("d", "e"),
					)),
				)),
			)
		}

		logger.Info("test array 2", logf.Object("obj", AsObject(logf.Array("aa", testArray{}))))

	} else {
		logger.Warn("couldn't get build info")
	}

	logger.Error("something bad happened", logf.Error(errors.New("failed to figure out what to do next")))
}

// ---

func AsObject(fields ...logf.Field) logf.ObjectEncoder {
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

type testArray struct{}

func (a testArray) EncodeLogfArray(enc logf.TypeEncoder) error {
	enc.EncodeTypeString("test")
	enc.EncodeTypeObject(AsObject(
		logf.Object("g", AsObject(
			logf.String("b", "c"),
		)),
		logf.String("d", "e"),
	))

	return nil
}

// ---

var _ logf.ArrayEncoder = testArray{}
