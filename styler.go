package logftxt

import (
	"github.com/ssgreg/logf"

	"github.com/pamburus/go-ansi-esc/sgr"
)

func newStyler() styler {
	return styler{
		defaultStyle,
		make(sgr.Sequence, 0, 8),
		false,
	}
}

type styler struct {
	style    style
	seq      sgr.Sequence
	disabled bool
}

func (s styler) Disabled(value bool) styler {
	s.disabled = value

	return s
}

func (s *styler) Use(style stylePatch, buf *logf.Buffer, f func()) {
	if s.disabled || style.IsEmpty {
		f()

		return
	}

	old := s.style
	if s.style.UpdateBy(style) {
		seq := old.diffToSequence(s.style, s.seq[0:0])
		buf.Data = seq.Render(buf.Data)
		f()
		seq = s.style.diffToSequence(old, s.seq[0:0])
		buf.Data = seq.Render(buf.Data)
		s.style = old
	} else {
		f()
	}
}

// ---

type style struct {
	Background sgr.Command
	Foreground sgr.Command
	Modes      sgr.ModeSet
}

type stylePatch struct {
	Background sgr.Command
	Foreground sgr.Command
	Modes      [4]sgr.ModeSet
	HasModes   bool
	IsEmpty    bool
}

func (s *style) UpdateBy(other stylePatch) bool {
	updated := false
	if !other.Background.IsZero() && other.Background != s.Background {
		s.Background = other.Background
		updated = true
	}
	if !other.Foreground.IsZero() && other.Foreground != s.Foreground {
		s.Foreground = other.Foreground
		updated = true
	}
	if other.HasModes {
		oldModes := s.Modes
		s.Modes = s.Modes.WithOther(other.Modes[sgr.ModeReplace], sgr.ModeAdd)
		s.Modes = s.Modes.WithOther(other.Modes[sgr.ModeAdd], sgr.ModeAdd)
		s.Modes = s.Modes.WithOther(other.Modes[sgr.ModeRemove], sgr.ModeRemove)
		s.Modes = s.Modes.WithOther(other.Modes[sgr.ModeToggle], sgr.ModeToggle)
		updated = updated || s.Modes != oldModes
	}

	return updated
}

func (s style) UpdatedBy(other stylePatch) style {
	s.UpdateBy(other)

	return s
}

func (s *style) diffToSequence(other style, seq sgr.Sequence) sgr.Sequence {
	if other == defaultStyle {
		return append(seq, sgr.ResetAll)
	}

	if other.Background != s.Background {
		seq = append(seq, other.Background)
	}
	if other.Foreground != s.Foreground {
		seq = append(seq, other.Foreground)
	}

	return s.Modes.Diff(other.Modes).ToCommands(seq)
}

// ---

var defaultStyle = style{
	Background: sgr.ResetBackgroundColor,
	Foreground: sgr.ResetForegroundColor,
	Modes:      sgr.NewModeSet(),
}
