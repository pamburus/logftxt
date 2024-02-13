package formatting_test

import (
	"testing"

	"github.com/pamburus/go-ansi-esc/sgr"
	"github.com/pamburus/go-tst/tst"
	"github.com/pamburus/logftxt/internal/pkg/themecfg/formatting"
)

func TestModePatch(tt *testing.T) {
	t := tst.New(tt)

	t.Run("MarshalText", func(t tst.Test) {
		t.Expect(
			formatting.ModePatch{Mode: sgr.Bold}.MarshalText(),
		).ToSucceed().AndResult().To(
			tst.Equal([]byte("Bold")),
		)

		t.Expect(
			formatting.ModePatch{Mode: sgr.Faint, Action: sgr.ModeAdd}.MarshalText(),
		).ToSucceed().AndResult().To(
			tst.Equal([]byte("+Faint")),
		)

		t.Expect(
			formatting.ModePatch{Mode: sgr.Framed, Action: sgr.ModeRemove}.MarshalText(),
		).ToSucceed().AndResult().To(
			tst.Equal([]byte("-Framed")),
		)

		t.Expect(
			formatting.ModePatch{Mode: sgr.CrossedOut, Action: sgr.ModeToggle}.MarshalText(),
		).ToSucceed().AndResult().To(
			tst.Equal([]byte("^CrossedOut")),
		)

		t.Expect(
			formatting.ModePatch{Mode: 227}.MarshalText(),
		).ToFailWith(
			sgr.ErrInvalidModeValue{Value: 227},
		)

		t.Expect(
			formatting.ModePatch{Mode: sgr.Bold, Action: 64}.MarshalText(),
		).ToFail()
	})

	t.Run("UnmarshalText", func(t tst.Test) {
		var p formatting.ModePatch

		t.Expect(p.UnmarshalText([]byte(""))).ToSucceed()
		t.Expect(p).To(tst.Equal(
			formatting.ModePatch{},
		))

		t.Expect(p.UnmarshalText([]byte("italic"))).ToSucceed()
		t.Expect(p).To(tst.Equal(
			formatting.ModePatch{Mode: sgr.Italic},
		))

		t.Expect(p.UnmarshalText([]byte("+Underlined"))).ToSucceed()
		t.Expect(p).To(tst.Equal(
			formatting.ModePatch{Mode: sgr.Underlined, Action: sgr.ModeAdd},
		))

		t.Expect(p.UnmarshalText([]byte("-CrossedOut"))).ToSucceed()
		t.Expect(p).To(tst.Equal(
			formatting.ModePatch{Mode: sgr.CrossedOut, Action: sgr.ModeRemove},
		))

		t.Expect(p.UnmarshalText([]byte("^rapid-blink"))).ToSucceed()
		t.Expect(p).To(tst.Equal(
			formatting.ModePatch{Mode: sgr.RapidBlink, Action: sgr.ModeToggle},
		))

		t.Expect(p.UnmarshalText([]byte("^rapid-blinking"))).ToFailWith(
			sgr.ErrInvalidModeText{Value: "rapid-blinking"},
		)
	})
}

func TestModePatchList(tt *testing.T) {
	t := tst.New(tt)

	pl := formatting.ModePatchList{
		{Mode: sgr.Bold},
		{Mode: sgr.Faint, Action: sgr.ModeAdd},
		{Mode: sgr.Framed, Action: sgr.ModeRemove},
		{Mode: sgr.Concealed, Action: sgr.ModeAdd},
		{Mode: sgr.CrossedOut, Action: sgr.ModeToggle},
		{Mode: sgr.Framed, Action: sgr.ModeReplace},
	}

	t.Expect(pl.Sets()).To(tst.Equal(
		[3]sgr.ModeSet{
			sgr.Bold.ModeSet() | sgr.Faint.ModeSet() | sgr.Concealed.ModeSet() | sgr.Framed.ModeSet(),
			0,
			sgr.CrossedOut.ModeSet(),
		},
	))
}

func TestStyle(tt *testing.T) {
	t := tst.New(tt)

	t.Run("UpdatedBy", func(t tst.Test) {
		t.Expect(
			formatting.Style{
				Foreground: sgr.Red.Color(),
				Background: sgr.Blue.Color(),
			}.UpdatedBy(
				formatting.Style{
					Background: sgr.Green.Color(),
				},
			),
		).To(tst.Equal(
			formatting.Style{
				Foreground: sgr.Red.Color(),
				Background: sgr.Green.Color(),
			},
		))

		t.Expect(
			formatting.Style{
				Foreground: sgr.Red.Color(),
				Background: sgr.Blue.Color(),
				Modes: formatting.ModePatchList{
					{Mode: sgr.Bold},
					{Mode: sgr.Faint, Action: sgr.ModeAdd},
					{Mode: sgr.Framed, Action: sgr.ModeReplace},
					{Mode: sgr.Concealed, Action: sgr.ModeAdd},
					{Mode: sgr.CrossedOut, Action: sgr.ModeToggle},
				},
			}.UpdatedBy(
				formatting.Style{
					Foreground: sgr.Green.Color(),
					Modes: formatting.ModePatchList{
						{Mode: sgr.Faint, Action: sgr.ModeRemove},
						{Mode: sgr.Framed, Action: sgr.ModeAdd},
						{Mode: sgr.CrossedOut, Action: sgr.ModeToggle},
					},
				},
			),
		).To(tst.Equal(
			formatting.Style{
				Foreground: sgr.Green.Color(),
				Background: sgr.Blue.Color(),
				Modes: formatting.ModePatchList{
					{Mode: sgr.Bold},
					{Mode: sgr.Faint, Action: sgr.ModeAdd},
					{Mode: sgr.Framed, Action: sgr.ModeReplace},
					{Mode: sgr.Concealed, Action: sgr.ModeAdd},
					{Mode: sgr.CrossedOut, Action: sgr.ModeToggle},
					{Mode: sgr.Faint, Action: sgr.ModeRemove},
					{Mode: sgr.Framed, Action: sgr.ModeAdd},
					{Mode: sgr.CrossedOut, Action: sgr.ModeToggle},
				},
			},
		))
	})
}

func TestFormat(tt *testing.T) {
	t := tst.New(tt)

	t.Run("UpdatedBy", func(t tst.Test) {
		t.Expect(
			formatting.Format{
				Prefix: "prefix",
				Suffix: "suffix",
				Style: formatting.Style{
					Foreground: sgr.Red.Color(),
					Background: sgr.Blue.Color(),
				},
			}.UpdatedBy(
				formatting.Format{
					Prefix: "new-prefix",
					Style: formatting.Style{
						Background: sgr.Green.Color(),
					},
				},
			),
		).To(tst.Equal(
			formatting.Format{
				Prefix: "new-prefix",
				Suffix: "suffix",
				Style: formatting.Style{
					Foreground: sgr.Red.Color(),
					Background: sgr.Green.Color(),
				},
			},
		))

		t.Expect(
			formatting.Format{
				Prefix: "prefix",
				Suffix: "suffix",
				Style: formatting.Style{
					Foreground: sgr.Red.Color(),
					Background: sgr.Blue.Color(),
				},
			}.UpdatedBy(
				formatting.Format{
					Suffix: "new-suffix",
				},
			),
		).To(tst.Equal(
			formatting.Format{
				Prefix: "prefix",
				Suffix: "new-suffix",
				Style: formatting.Style{
					Foreground: sgr.Red.Color(),
					Background: sgr.Blue.Color(),
				},
			},
		))
	})
}

func TestItem(tt *testing.T) {
	t := tst.New(tt)

	t.Run("IsZero", func(t tst.Test) {
		t.Expect(
			formatting.Style{}.IsZero(),
		).To(tst.BeTrue())

		t.Expect(
			formatting.Style{
				Background: sgr.Red.Color(),
			}.IsZero(),
		).To(tst.BeFalse())

		t.Expect(
			formatting.Style{
				Foreground: sgr.Red.Color(),
			}.IsZero(),
		).To(tst.BeFalse())

		t.Expect(
			formatting.Style{
				Modes: formatting.ModePatchList{
					{Mode: sgr.Bold},
				},
			}.IsZero(),
		).To(tst.BeFalse())
	})

	t.Run("UpdatedBy", func(t tst.Test) {
		t.Expect(
			formatting.Item{
				Outer: formatting.Format{
					Prefix: "prefix",
					Suffix: "suffix",
					Style: formatting.Style{
						Foreground: sgr.Red.Color(),
					},
				},
				Separator: formatting.StyledText{
					Text: "separator",
					Style: formatting.Style{
						Foreground: sgr.Red.Color(),
						Background: sgr.Blue.Color(),
					},
				},
				Text: "text",
			}.UpdatedBy(
				formatting.Item{
					Outer: formatting.Format{
						Prefix: "new-prefix",
						Style: formatting.Style{
							Background: sgr.Green.Color(),
						},
					},
					Inner: formatting.Format{
						Prefix: "inner-prefix",
						Suffix: "inner-suffix",
						Style: formatting.Style{
							Foreground: sgr.Yellow.Color(),
							Background: sgr.Magenta.Color(),
						},
					},
					Separator: formatting.StyledText{
						Text: "new-separator",
						Style: formatting.Style{
							Foreground: sgr.Yellow.Color(),
						},
					},
					Text: "new-text",
				},
			),
		).To(tst.Equal(
			formatting.Item{
				Outer: formatting.Format{
					Prefix: "new-prefix",
					Suffix: "suffix",
					Style: formatting.Style{
						Foreground: sgr.Red.Color(),
						Background: sgr.Green.Color(),
					},
				},
				Inner: formatting.Format{
					Prefix: "inner-prefix",
					Suffix: "inner-suffix",
					Style: formatting.Style{
						Foreground: sgr.Yellow.Color(),
						Background: sgr.Magenta.Color(),
					},
				},
				Separator: formatting.StyledText{
					Text: "new-separator",
					Style: formatting.Style{
						Foreground: sgr.Yellow.Color(),
						Background: sgr.Blue.Color(),
					},
				},
				Text: "new-text",
			},
		))
	})
}
