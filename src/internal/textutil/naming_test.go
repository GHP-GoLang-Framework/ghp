package textutil

import "testing"

func TestPascalCase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"already Pascal", "About", "About"},
		{"kebab-case", "my-page", "MyPage"},
		{"underscore", "faq_page", "FaqPage"},
		{"mixed separators", "faq-page_v2", "FaqPageV2"},
		{"accented letter", "órgão", "Órgão"},
		{"spaces", "hello world", "HelloWorld"},
		{"leading separator", "-slug", "Slug"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PascalCase(tt.in); got != tt.want {
				t.Errorf("PascalCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeASCII(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"accented vowels", "órgão", "orgao"},
		{"plain ascii", "page", "page"},
		{"multiple accents", "àáâãéêíóôõú", "aaaaeeiooou"},
		{"uppercase accents", "ÀÉÕ", "AEO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeASCII(tt.in); got != tt.want {
				t.Errorf("NormalizeASCII(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStripNonWord(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"brackets", "[slug]", "slug"},
		{"dots", "a.b", "ab"},
		{"keeps alnum", "Abc123", "Abc123"},
		{"removes spaces", "a b", "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripNonWord(tt.in); got != tt.want {
				t.Errorf("StripNonWord(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBracketsToBraces(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"single segment", "[id]", "{id}"},
		{"path segment", "a/[id]-b", "a/{id}-b"},
		{"no brackets", "plain", "plain"},
		{"kebab slug", "[my-slug]", "{my-slug}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BracketsToBraces(tt.in); got != tt.want {
				t.Errorf("BracketsToBraces(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
