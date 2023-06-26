package logftxt_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt"
)

func TestErrorShort(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(
		string(logftxt.ErrorShort()(nil, mockError{})),
	).ToEqual(
		mockError{}.Error(),
	)

	t.Expect(
		string(logftxt.ErrorLong()(nil, mockError{})),
	).ToEqual(
		mockDetailedErrorMessage,
	)
}

// ---

type mockError struct{}

func (mockError) Error() string {
	return "mock error"
}

func (e mockError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = s.Write([]byte(mockDetailedErrorMessage))

			return
		}

		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

// ---

const mockDetailedErrorMessage = "detailed error"
