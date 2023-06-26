package pathx_test

import (
	"testing"

	"github.com/pamburus/go-tst/tst"

	"github.com/pamburus/logftxt/internal/pkg/pathx"
)

func TestCutPrefix(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.CutPrefix("", "")).ToEqual("", true)
	t.Expect(pathx.CutPrefix("", "a")).ToEqual("", false)
	t.Expect(pathx.CutPrefix("a", "")).ToEqual("a", true)
	t.Expect(pathx.CutPrefix("a", "a")).ToEqual(".", true)
	t.Expect(pathx.CutPrefix("a", "a/")).ToEqual("", false)
	t.Expect(pathx.CutPrefix("a/", "a")).ToEqual(".", true)
	t.Expect(pathx.CutPrefix("a/b", "a")).ToEqual("b", true)
	t.Expect(pathx.CutPrefix("a/b", "a/")).ToEqual("b", true)
	t.Expect(pathx.CutPrefix("a/", "a/")).ToEqual(".", true)
	t.Expect(pathx.CutPrefix("a/bb", "a/b")).ToEqual("", false)
}

func TestCutSuffix(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.CutSuffix("", "")).ToEqual("", true)
	t.Expect(pathx.CutSuffix("", "a")).ToEqual("", false)
	t.Expect(pathx.CutSuffix("a", "")).ToEqual("a", true)
	t.Expect(pathx.CutSuffix("a", "a")).ToEqual(".", true)
	t.Expect(pathx.CutSuffix("a", "a/")).ToEqual("", false)
	t.Expect(pathx.CutSuffix("a/", "a")).ToEqual("", false)
	t.Expect(pathx.CutSuffix("a/b", "b")).ToEqual("a", true)
	t.Expect(pathx.CutSuffix("a/b", "/b")).ToEqual("", false)
	t.Expect(pathx.CutSuffix("a/b", "b")).ToEqual("a", true)
	t.Expect(pathx.CutSuffix("a/", "a/")).ToEqual(".", true)
	t.Expect(pathx.CutSuffix("aa/b", "a/b")).ToEqual("", false)
	t.Expect(pathx.CutSuffix("a/b", "a/b")).ToEqual(".", true)
}

func TestHasPrefix(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.HasPrefix("a/b", "a")).ToBeTrue()
	t.Expect(pathx.HasPrefix("aa/b", "a")).ToBeFalse()
}

func TestHasSuffix(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.HasSuffix("a/b", "b")).ToBeTrue()
	t.Expect(pathx.HasSuffix("a/bb", "b")).ToBeFalse()
}

func TestExplicitlyRelative(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.ExplicitlyRelative("/a/b")).ToEqual("/a/b")
	t.Expect(pathx.ExplicitlyRelative("a/b")).ToEqual("./a/b")
	t.Expect(pathx.ExplicitlyRelative("./a/b")).ToEqual("./a/b")
	t.Expect(pathx.ExplicitlyRelative("../a/b")).ToEqual("../a/b")
}

func TestOS(tt *testing.T) {
	t := tst.New(tt)

	t.Expect(pathx.OS().ExplicitlyRelative("a")).To(tst.Or(tst.Equal("./a"), tst.Equal(".\\a")))
}
