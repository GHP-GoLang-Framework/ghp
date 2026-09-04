// Package colors wraps strings with ANSI escape codes for terminal colors,
// appending the reset code after the text.
package colors

import "strings"

const reset = "\033[0m"

// Green wraps s with the ANSI green code and a trailing reset.
func Green(s string) string {
	return "\033[32m" + s + reset
}

// Red wraps s with the ANSI red code and a trailing reset.
func Red(s string) string {
	return "\033[31m" + s + reset
}

// Yellow wraps s with the ANSI yellow code and a trailing reset.
func Yellow(s string) string {
	return "\033[33m" + s + reset
}

// Blue wraps s in ANSI blue and a reset.
func Blue(s string) string {
	return "\033[34m" + s + reset
}

// Magenta wraps s in ANSI magenta and a reset.
func Magenta(s string) string {
	return "\033[35m" + s + reset
}

// Cyan wraps s in ANSI cyan and a reset.
func Cyan(s string) string {
	return "\033[36m" + s + reset
}

// Pad left-justifies s to width n counting only visible characters,
// so ANSI escapes never shift the column. Ex: Pad("\033[32mOK\033[0m", 4)
// -> "\033[32mOK\033[0m  ".
func Pad(s string, n int) string {
	w := len(stripANSI(s))
	if w >= n {
		return s
	}
	return s + spaces(n-w)
}

// Dots left-justifies s to width n using a dotted fill, counting only
// visible characters so ANSI escapes never shift the column.
// Ex: Dots("OK", 6) -> "OK....".
func Dots(s string, n int) string {
	w := len(stripANSI(s))
	if w >= n {
		return s
	}
	return s + strings.Repeat(".", n-w)
}

// stripANSI removes ANSI escape sequences, returning the visible width.
func stripANSI(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			for i += 2; i < len(s) && !((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z')); i++ {
			}
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

func spaces(n int) string {
	return strings.Repeat(" ", n)
}
