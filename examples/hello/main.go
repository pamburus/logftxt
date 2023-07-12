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
		Appender: logftxt.NewAppender(os.Stdout),
	})
	defer writerClose()

	// Create Logger with ChannelWriter.
	logger := logf.NewLogger(logf.LevelDebug, writer).WithCaller().WithName("main")

	// Do some logging.
	logger.Info("runtime info", logf.Int("cpu-count", runtime.NumCPU()))
	logger.Info("test array", logf.Any("ss", []string{"cpu-count", "abc"}))

	if info, ok := debug.ReadBuildInfo(); ok {
		logger.Debug("build info", logf.String("go-version", info.GoVersion), logf.String("path", info.Path))
		for _, setting := range info.Settings {
			logger.Debug("build setting", logf.String("key", setting.Key), logf.String("value", setting.Value))
		}
		for _, dep := range info.Deps {
			logger.Debug("dependency", logf.String("path", dep.Path), logf.String("version", dep.Version))
		}
	} else {
		logger.Warn("couldn't get build info")
	}

	logger.Error("something bad happened", logf.Error(errors.New("failed to figure out what to do next")))
}
