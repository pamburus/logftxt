package logftxt

import (
	"os"
	"testing"

	"github.com/pamburus/go-tst/tst"
)

func TestError(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(ErrFileNotFound{}).ToNot(tst.MatchError(os.ErrNotExist))
	t.Expect(ErrFileNotFound{}.Error()).ToNot(tst.BeZero())
}
