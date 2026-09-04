package pages

import "testing"

func TestNewPage(t *testing.T) {
	tests := []struct {
		name       string
		relPath    string
		wantRelDir string
		wantFile   string
		wantFunc   string
		wantPkg    string
	}{
		{
			name:       "top level page",
			relPath:    "about.ghp",
			wantRelDir: "",
			wantFile:   "about",
			wantFunc:   "About",
			wantPkg:    "main",
		},
		{
			name:       "nested page",
			relPath:    "blog/about.ghp",
			wantRelDir: "blog",
			wantFile:   "about",
			wantFunc:   "About",
			wantPkg:    "blog",
		},
		{
			name:       "deeply nested page",
			relPath:    "blog/tech/go.ghp",
			wantRelDir: "blog/tech",
			wantFile:   "go",
			wantFunc:   "Go",
			wantPkg:    "tech",
		},
		{
			name:       "separators stripped from func name",
			relPath:    "docs/faq-page_v2.ghp",
			wantRelDir: "docs",
			wantFile:   "faq-page_v2",
			wantFunc:   "FaqPageV2",
			wantPkg:    "docs",
		},
		{
			name:       "toolbar-i18n matches docs template func name",
			relPath:    "user/toolbar-i18n.ghp",
			wantRelDir: "user",
			wantFile:   "toolbar-i18n",
			wantFunc:   "ToolbarI18n",
			wantPkg:    "user",
		},
		{
			name:       "uppercase dir lowercased as pkg",
			relPath:    "Docs/Faq.ghp",
			wantRelDir: "Docs",
			wantFile:   "Faq",
			wantFunc:   "Faq",
			wantPkg:    "docs",
		},
		{
			name:       "accents stripped",
			relPath:    "orgao/órgão.ghp",
			wantRelDir: "orgao",
			wantFile:   "órgão",
			wantFunc:   "Orgao",
			wantPkg:    "orgao",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPage(tt.relPath)

			if got.RelDir != tt.wantRelDir {
				t.Errorf("RelDir = %q, want %q", got.RelDir, tt.wantRelDir)
			}
			if got.FileName != tt.wantFile {
				t.Errorf("FileName = %q, want %q", got.FileName, tt.wantFile)
			}
			if got.FuncName != tt.wantFunc {
				t.Errorf("FuncName = %q, want %q", got.FuncName, tt.wantFunc)
			}
			if got.PkgName != tt.wantPkg {
				t.Errorf("PkgName = %q, want %q", got.PkgName, tt.wantPkg)
			}
		})
	}
}
