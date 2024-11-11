package logftxt

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/pamburus/go-tst/tst"
)

func TestConfig(tt *testing.T) {
	t := tst.New(tt)

	t.Run("Load", func(t tst.Test) {
		envConfig := func(value string) Environment {
			return func(name string) (string, bool) {
				switch name {
				case "LOGFTXT_CONFIG":
					return value, true
				}

				return "", false
			}
		}

		_ = DefaultConfig()
		t.Expect(LoadConfig("./assets/config.yml")).ToSucceed()
		t.Expect(LoadConfig("./_.yml")).ToFailWith(ErrFileNotFound{Filename: "./_.yml"})
		t.Expect(LoadConfig("./_.yml")).ToFailWith(ErrFileNotFound{})
		t.Expect(LoadConfig("./config.go")).ToFail()

		t.Expect(ConfigFromDefaultPath(WithFS(mockFS{openErr: errOpen}))(domain{})).ToFailWith(errOpen)
		t.Expect(ConfigFromDefaultPath()(domain{fs: mockFS{openErr: errOpen}})).ToFailWith(errOpen)
		t.Expect(ConfigFromDefaultPath(WithFS(mockFS{openErr: ErrFileNotFound{}}))(domain{})).ToFailWith(ErrFileNotFound{})
		t.Expect(ConfigFromDefaultPath()(domain{fs: mockFS{openErr: ErrFileNotFound{}}})).ToFailWith(ErrFileNotFound{})
		t.Expect(ConfigFromDefaultPath(
			WithFS(mockFS{file: &mockFile{bytes.NewBuffer(mockConfigData)}}))(domain{}),
		).ToSucceed()
		t.Expect(
			ConfigFromEnvironment(
				envConfig("a.yml"),
				WithFS(mockFS{file: &mockFile{bytes.NewBuffer(mockConfigData)}}),
			)(domain{}),
		).ToSucceed()
	})

	t.Run("Read", func(t tst.Test) {
		t.Run("Empty", func(t tst.Test) {
			t.Expect(
				ReadConfig(strings.NewReader(`{}`)),
			).ToSucceed()
		})
		t.Run("InvalidDurationFormat", func(t tst.Test) {
			t.Expect(
				ReadConfig(strings.NewReader(`{"values": {"duration": {"format": "aaa"}}}`)),
			).ToFail()
			t.Expect(
				ReadConfig(strings.NewReader(`{"values": {"error": {"format": "aaa"}}}`)),
			).ToFail()
			t.Expect(
				ReadConfig(strings.NewReader(`{"caller": {"format": "aaa"}}`)),
			).ToFail()
		})
	})

	t.Run("Options", func(t tst.Test) {
		t.Expect(encoderOptions{}.With([]EncoderOption{Config{}}).provideConfig).ToNot(tst.BeZero())
		t.Expect(appenderOptions{}.With([]AppenderOption{Config{}}).provideConfig).ToNot(tst.BeZero())
	})
}

// ---

type mockFS struct {
	dirErr  error
	openErr error
	file    fs.File
}

func (f mockFS) ConfigDir() (fs.FS, error) {
	return f, f.dirErr
}

func (f mockFS) Open(string) (fs.File, error) {
	return f.file, f.openErr
}

// ---

type mockFile struct {
	io.Reader
}

func (f mockFile) Stat() (fs.FileInfo, error) {
	return nil, errNotImplemented
}

func (f mockFile) Close() error {
	return nil
}

// ---

var (
	errOpen           = errors.New("mock open err")
	errNotImplemented = errors.New("not implemented")
)

//go:embed assets/config.yml
var mockConfigData []byte
