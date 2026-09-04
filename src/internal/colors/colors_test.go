package colors

import "testing"

func TestWrap(t *testing.T) {
	tests := []struct {
		name string
		wrap func(string) string
		code string
	}{
		{"Green", Green, "\033[32m"},
		{"Red", Red, "\033[31m"},
		{"Yellow", Yellow, "\033[33m"},
		{"Blue", Blue, "\033[34m"},
		{"Magenta", Magenta, "\033[35m"},
		{"Cyan", Cyan, "\033[36m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.code + "x" + reset
			if got := tt.wrap("x"); got != want {
				t.Errorf("%s(\"x\") = %q, want %q", tt.name, got, want)
			}
		})
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"plain", "OK", 4, "OK  "},
		{"counting ansi as invisible", Green("OK"), 4, Green("OK") + "  "},
		{"longer than width returns unchanged", "too long", 4, "too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Pad(tt.in, tt.n); got != tt.want {
				t.Errorf("Pad(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
			}
		})
	}
}

func TestDots(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"plain", "OK", 6, "OK...."},
		{"counting ansi as invisible", Green("OK"), 6, Green("OK") + "...."},
		{"longer than width returns unchanged", "too long", 4, "too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dots(tt.in, tt.n); got != tt.want {
				t.Errorf("Dots(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
			}
		})
	}
}
