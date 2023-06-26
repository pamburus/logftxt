package themecfg

import "github.com/pamburus/logftxt/internal/pkg/themecfg/formatting"

// Style includes background color, foreground color and a set of modes.
// Modes overwrites current set of modes during style rendering.
// So, explicitly specifying empty list of modes will disable all currently enabled modes.
type Style = formatting.Style
