package convention

import (
	"fmt"
	"regexp"
)

// Parse compiles a ConventionConfig into a usable Convention.
// Returns an error if the pattern is empty or invalid.
func Parse(cfg ConventionConfig) (*Convention, error) {
	if cfg.Pattern == "" {
		return nil, fmt.Errorf("convention %q: pattern is required", cfg.Name)
	}

	// ScopeDelimiter 허용값 검증
	switch cfg.ScopeDelimiter {
	case "", "()", "[]":
		// 허용
	default:
		return nil, fmt.Errorf("convention %q: invalid scope_delimiter %q: must be \"\", \"()\", or \"[]\"", cfg.Name, cfg.ScopeDelimiter)
	}

	compiled, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("convention %q: invalid pattern: %w", cfg.Name, err)
	}

	maxLen := cfg.MaxLength
	if maxLen <= 0 {
		maxLen = 100 // default
	}

	return &Convention{
		Name:           cfg.Name,
		Pattern:        compiled,
		Types:          cfg.Types,
		Scopes:         cfg.Scopes,
		MaxLength:      maxLen,
		Required:       cfg.Required,
		Examples:       cfg.Examples,
		ScopeDelimiter: cfg.ScopeDelimiter,
	}, nil
}

// ParseBuiltin loads and compiles a built-in convention by name.
// Returns an error if the name is not recognized.
func ParseBuiltin(name string) (*Convention, error) {
	cfg := GetBuiltin(name)
	if cfg == nil {
		return nil, fmt.Errorf("unknown built-in convention: %q", name)
	}
	return Parse(*cfg)
}
