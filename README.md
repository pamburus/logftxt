# logftxt [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Package `logftxt` provides an alternative [logf](github.com/ssgreg/logf) Appender and Encoder for colored text logs.
It can be used as a more flexible replacement for [logftext](https://github.com/ssgreg/logftext).

## Features

* Highlighting of log levels, messages, field names, delimiters, arrays, objects, strings, numbers, errors, durations, etc.
* Built-in themes and external [custom](#changing-theme) themes.
* Built-in [@default](assets/theme/default.yml) and [@fancy](assets/theme/fancy.yml) themes have good support for both [dark](examples/hello/assets/screenshots/hello-dark-fancy.png) and [light](examples/hello/assets/screenshots/hello-light-fancy.png) terminals.
* Built-in [configuration file](assets/config.yml) that can easily be replaced with a [custom](#changing-configuration) configuration file.

### Changing configuration
    
Built-in [configuration file](assets/config.yml) can be easily replaced with custom configuration file by
* Copying it, editing it, and setting up an environment variable `LOGFTXT_CONFIG` pointing to the file
    * `path/to/my-config.yml` for a custom config relative to `~/.config/logftxt`
    * `./path/to/my-config.yml` for a custom config relative to the current directory
    * `/home/root/path/to/my-config.yml` for a custom config at an absolute path
* Loading it manually with `LoadConfig` or `ReadConfig` and passing it as an optional parameter to `NewAppender` or `NewEncoder` function

### Changing theme

Theme can be easily changed by
* Setting up `LOGFTXT_THEME` environment variable to a value of
    * `@default`, `@fancy` or `@legacy` for a built-in theme
    * `path/to/my-theme.yml` for a custom theme relative to `~/.config/logftxt`
    * `./path/to/my-theme.yml` for a custom theme relative to the current directory
    * `/home/root/path/to/my-theme.yml` for a custom theme at an absolute path
* Setting the theme name in the custom configuration file in the same format as the `LOGFTXT_THEME` environment variable
* Providing an optional parameter `ThemeRef` to `NewAppender` or `NewEncoder` function containing the same value as the `LOGFTXT_THEME` environment variable
* Loading it manually with `LoadTheme` or `ReadTheme` and passing it as an optional parameter to `NewAppender` or `NewEncoder` function

## Example

The following example creates the new `logf` logger with the `logftxt` Appender constructed with the default Encoder.
The source code can be found in [examples/hello](examples/hello/main.go).

```go
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
```

### Example output
![GitHub-Mark-Light](examples/hello/assets/screenshots/hello-light-fancy.png#gh-light-mode-only)
![GitHub-Mark-Dark ](examples/hello/assets/screenshots/hello-dark-fancy.png#gh-dark-mode-only)


### Used terminal color schemes

#### iTerm2
* [One Dark Neo](https://gist.github.com/pamburus/0ad130f2af9ab03a97f2a9f7b4f18c68/746ca7103726d43b767f2111799d3cb5ec08adbb)
* Built-in "Light Background" color scheme

#### Alacritty
* [One Dark Neo](https://gist.github.com/pamburus/e27ebf60aa17d126f5c879f06112edd6/a1e66d34a65b883f1cb8ec28820cc0c53233e3aa#file-alacritty-yml-L904)
  * Note: It is recommended to use `draw_bold_text_with_bright_colors: true` setting
* [Light](https://gist.github.com/pamburus/e27ebf60aa17d126f5c879f06112edd6/a1e66d34a65b883f1cb8ec28820cc0c53233e3aa#file-alacritty-yml-L875)
  * Note: It is recommended to use `draw_bold_text_with_bright_colors: false` setting


[doc-img]: https://pkg.go.dev/badge/github.com/pamburus/logftxt
[doc]: https://pkg.go.dev/github.com/pamburus/logftxt
[ci-img]: https://github.com/pamburus/logftxt/actions/workflows/ci.yml/badge.svg
[ci]: https://github.com/pamburus/logftxt/actions/workflows/ci.yml
[cov-img]: https://codecov.io/gh/pamburus/logftxt/logftxt/branch/main/graph/badge.svg?token=YHWG3WZG7C
[cov]: https://codecov.io/gh/pamburus/logftxt/logftxt
