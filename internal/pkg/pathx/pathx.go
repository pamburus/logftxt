// Package pathx provides additional utilities for working with paths.
package pathx

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ---

// OS returns current operating system specific setup.
func OS() Setup {
	return Setup{os.PathSeparator, filepath.IsAbs}
}

// Posix returns operating system independent setup conforming `posix` standards.
func Posix() Setup {
	return Setup{'/', path.IsAbs}
}

// HasPrefix delegates the call to Setup.HasPrefix assuming Posix setup.
func HasPrefix(path, prefix string) bool {
	return Posix().HasPrefix(path, prefix)
}

// HasSuffix delegates the call to Setup.HasSuffix assuming Posix setup.
func HasSuffix(path, suffix string) bool {
	return Posix().HasSuffix(path, suffix)
}

// CutPrefix delegates the call to Setup.CutPrefix assuming Posix setup.
func CutPrefix(path, prefix string) (string, bool) {
	return Posix().CutPrefix(path, prefix)
}

// CutSuffix delegates the call to Setup.CutSuffix assuming Posix setup.
func CutSuffix(path, suffix string) (string, bool) {
	return Posix().CutSuffix(path, suffix)
}

// ExplicitlyRelative delegates the call to Setup.ExplicitlyRelative assuming Posix setup.
func ExplicitlyRelative(p string) string {
	return Posix().ExplicitlyRelative(p)
}

// ---

// Setup is a certain configuration of paths handling rules.
type Setup struct {
	separator byte
	isAbs     func(string) bool
}

// HasPrefix checks whether path has the given prefix.
// Path is considered to have a prefix only if the prefix hits at path delimiter boundary.
func (s Setup) HasPrefix(path, prefix string) bool {
	_, ok := s.CutPrefix(path, prefix)

	return ok
}

// HasSuffix checks whether path has the given suffix.
// Path is considered to have a suffix only if the suffix hits at path delimiter boundary.
func (s Setup) HasSuffix(path, suffix string) bool {
	_, ok := s.CutSuffix(path, suffix)

	return ok
}

// CutPrefix checks whether path has the given prefix and returns remaining path if it has.
// Path is considered to have a prefix only if the prefix hits at path delimiter boundary.
func (s Setup) CutPrefix(path, prefix string) (string, bool) {
	if len(prefix) == 0 {
		return path, true
	}

	part, ok := strings.CutPrefix(path, prefix)
	if !ok || (len(part) != 0 && !s.isSeparator(part[0]) && !s.isSeparator(prefix[len(prefix)-1])) {
		return "", false
	}

	part = trimLeft(part, s.isSeparator)
	if part == "" {
		part = "."
	}

	return part, true
}

// CutSuffix checks whether path has the given suffix and returns remaining path if it has.
// Path is considered to have a suffix only if the suffix hits at path delimiter boundary.
func (s Setup) CutSuffix(path, suffix string) (string, bool) {
	if len(suffix) == 0 {
		return path, true
	}

	if s.isSeparator(suffix[0]) {
		return "", false
	}

	part, ok := strings.CutSuffix(path, suffix)
	if !ok || (len(part) != 0 && !s.isSeparator(part[len(part)-1])) {
		return "", false
	}

	part = trimRight(part, s.isSeparator)
	if part == "" {
		part = "."
	}

	return part, true
}

// ExplicitlyRelative checks if prefix is a relative path and ensures it has '.' or '..' prefix.
// If it doesn't, '.' prefix is added. If it is absolute, original value is returned.
func (s Setup) ExplicitlyRelative(path string) string {
	if s.isAbs(path) {
		return path
	}

	if s.HasPrefix(path, ".") || s.HasPrefix(path, "..") {
		return path
	}

	return fmt.Sprintf(".%c%s", s.separator, path)
}

func (s Setup) isSeparator(c byte) bool {
	return c == s.separator
}

// ---

func trimLeft(s string, f func(byte) bool) string {
	if i, ok := findByte(s, not(f)); ok {
		return s[i:]
	}

	return ""
}

func trimRight(s string, f func(byte) bool) string {
	if i, ok := findLastByte(s, not(f)); ok {
		return s[:i+1]
	}

	return ""
}

func findByte(s string, f func(byte) bool) (int, bool) {
	for i := 0; i != len(s); i++ {
		if f(s[i]) {
			return i, true
		}
	}

	return -1, false
}

func findLastByte(s string, f func(byte) bool) (int, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if f(s[i]) {
			return i, true
		}
	}

	return -1, false
}

func not(f func(byte) bool) func(byte) bool {
	return func(b byte) bool {
		return !f(b)
	}
}
