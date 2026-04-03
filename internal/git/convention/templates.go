package convention

import "sort"

// builtinConventions holds the built-in convention configurations.
var builtinConventions = map[string]ConventionConfig{
	"conventional-commits": {
		Name:           "conventional-commits",
		Pattern:        `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\(.+\))?!?: .+`,
		Types:          []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
		MaxLength:      100,
		ScopeDelimiter: "()",
		Examples: []string{
			"feat(auth): add JWT token validation",
			"fix: resolve null pointer in user service",
			"docs(readme): update installation guide",
		},
	},
	"angular": {
		Name:           "angular",
		Pattern:        `^(build|ci|docs|feat|fix|perf|refactor|test)(\([a-z-]+\))?: .+`,
		Types:          []string{"build", "ci", "docs", "feat", "fix", "perf", "refactor", "test"},
		MaxLength:      100,
		ScopeDelimiter: "()",
		Examples: []string{
			"feat(router): add lazy loading support",
			"fix(compiler): handle edge case in template parser",
		},
	},
	"karma": {
		Name:           "karma",
		Pattern:        `^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?: .+`,
		Types:          []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "chore"},
		MaxLength:      100,
		ScopeDelimiter: "()",
		Examples: []string{
			"feat(service): add user notification endpoint",
			"test(api): add integration tests for auth module",
		},
	},
	"bracket-scope": {
		Name:           "bracket-scope",
		Pattern:        `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)!?: (\[[A-Za-z][\w]*(/[A-Za-z][\w]*)*\] )?[^\[\]].+`,
		Types:          []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
		MaxLength:      100,
		ScopeDelimiter: "[]",
		Examples: []string{
			"feat: [Web] add restore button with config option",
			"fix: [Auth] resolve token expiration issue",
			"refactor: [Core/DB] optimize query performance",
			"feat!: [API] change login endpoint signature",
			"chore: [Build] update dependencies",
		},
	},
}

// BuiltinNames returns the list of available built-in convention names.
// 정렬된 슬라이스를 반환한다.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtinConventions))
	for k := range builtinConventions {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// GetBuiltin returns a built-in convention config by name.
// Returns nil if not found.
func GetBuiltin(name string) *ConventionConfig {
	cfg, ok := builtinConventions[name]
	if !ok {
		return nil
	}
	return &cfg
}
