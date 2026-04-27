package template

import (
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// lspServerEntry mirrors the YAML schema for a single language server entry
// used in lsp.yaml.tmpl. This struct is intentionally local to the test to
// validate that the template can be round-tripped through a typed struct.
type lspServerEntry struct {
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
	FileExtensions []string `yaml:"file_extensions"`
}

// lspConfig mirrors the top-level structure of lsp.yaml.tmpl.
type lspConfig struct {
	LSP struct {
		Servers map[string]lspServerEntry `yaml:"servers"`
	} `yaml:"lsp"`
}

// readLSPTemplate loads lsp.yaml.tmpl from the embedded template FS.
func readLSPTemplate(t *testing.T) []byte {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".ae/config/sections/lsp.yaml.tmpl")
	if err != nil {
		t.Fatalf("read lsp.yaml.tmpl: %v", err)
	}
	return data
}

// TestLSPSchemaKeyCommand asserts that:
//   - No entry uses the deprecated "binary:" key (raw text — schema validity)
//   - At least 16 servers exist with non-empty "command:" field (parsed YAML)
//
// The "command:" field count is enforced via yaml.Unmarshal rather than raw
// string counting so the test is independent of indentation/formatting and
// remains stable when the template is reformatted.
func TestLSPSchemaKeyCommand(t *testing.T) {
	t.Parallel()

	data := readLSPTemplate(t)

	// Schema validity: no deprecated "binary:" key anywhere in raw template.
	// Raw-text scan catches accidental reintroduction independent of YAML structure.
	if strings.Contains(string(data), "binary:") {
		t.Error("lsp.yaml.tmpl contains deprecated 'binary:' key; use 'command:' instead")
	}

	// Structural validity: parse YAML and verify servers have non-empty Command.
	var cfg lspConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("yaml.Unmarshal(lsp.yaml.tmpl): %v", err)
	}

	const wantMinLanguages = 16
	commandCount := 0
	for lang, entry := range cfg.LSP.Servers {
		if entry.Command == "" {
			t.Errorf("language %q has empty 'command' field", lang)
			continue
		}
		commandCount++
	}
	if commandCount < wantMinLanguages {
		t.Errorf("lsp.yaml.tmpl has %d servers with non-empty command; want >= %d (one per supported language)",
			commandCount, wantMinLanguages)
	}
}

// TestLSPFileExtensionsComplete asserts that every language entry in
// lsp.yaml.tmpl has a non-empty file_extensions list.
func TestLSPFileExtensionsComplete(t *testing.T) {
	t.Parallel()

	data := readLSPTemplate(t)

	var cfg lspConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("yaml.Unmarshal(lsp.yaml.tmpl): %v", err)
	}

	if len(cfg.LSP.Servers) == 0 {
		t.Fatal("lsp.yaml.tmpl parsed to zero language servers; check YAML structure")
	}

	const wantMinLanguages = 16
	if len(cfg.LSP.Servers) < wantMinLanguages {
		t.Errorf("expected at least %d language servers, got %d", wantMinLanguages, len(cfg.LSP.Servers))
	}

	for lang, entry := range cfg.LSP.Servers {
		if len(entry.FileExtensions) == 0 {
			t.Errorf("language %q has empty file_extensions list", lang)
		}
	}
}

// TestLSPStructTagConsistency verifies that lsp.yaml.tmpl can be
// unmarshalled into the lspConfig struct and then re-marshalled to YAML,
// producing output that retains all 16 language server keys.
// This acts as a YAML round-trip test against the typed ServerConfig schema.
func TestLSPStructTagConsistency(t *testing.T) {
	t.Parallel()

	data := readLSPTemplate(t)

	var cfg lspConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal lsp.yaml.tmpl into lspConfig: %v", err)
	}

	// Re-marshal and unmarshal to verify round-trip stability.
	out, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal(lspConfig): %v", err)
	}

	var cfg2 lspConfig
	if err := yaml.Unmarshal(out, &cfg2); err != nil {
		t.Fatalf("yaml.Unmarshal(round-trip output): %v", err)
	}

	// Verify all original languages survive the round-trip.
	for lang := range cfg.LSP.Servers {
		if _, ok := cfg2.LSP.Servers[lang]; !ok {
			t.Errorf("language %q lost during YAML round-trip", lang)
		}
	}

	// Verify mandatory languages are present.
	mandatoryLanguages := []string{
		"go", "python", "typescript", "javascript", "rust", "java",
		"kotlin", "swift", "ruby", "php", "cpp", "csharp",
		"scala", "elixir", "r", "dart",
	}
	for _, lang := range mandatoryLanguages {
		if _, ok := cfg2.LSP.Servers[lang]; !ok {
			t.Errorf("mandatory language %q missing after round-trip", lang)
		}
	}

	// Verify "dart" is used (NOT "flutter").
	if _, ok := cfg2.LSP.Servers["flutter"]; ok {
		t.Error("found 'flutter' key; must use 'dart' per v1.2.0 specification")
	}

	// Verify every server has a non-empty command.
	for lang, entry := range cfg2.LSP.Servers {
		if entry.Command == "" {
			t.Errorf("language %q has empty command field after round-trip", lang)
		}
	}
}
