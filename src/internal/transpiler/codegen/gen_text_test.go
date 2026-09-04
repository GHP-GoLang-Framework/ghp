package codegen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenText(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "plain text",
			value: "hello world",
			want:  "io.WriteString(w, \"hello world\")\n",
		},
		{
			name:  "strips line breaks and tabs",
			value: "a\n\tb",
			want:  "io.WriteString(w, \"ab\")\n",
		},
		{
			name:  "only whitespace after stripping is dropped",
			value: "\n\t  \n",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			genText(&b, ast.NewText(tt.value, 1))

			if got := b.String(); got != tt.want {
				t.Errorf("genText() = %q, want %q", got, tt.want)
			}
		})
	}
}
