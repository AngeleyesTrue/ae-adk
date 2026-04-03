package convention

import (
	"fmt"
	"slices"
	"strings"
)

// Validate checks a commit message against a convention.
// If conv is nil the message is considered valid.
func Validate(message string, conv *Convention) ValidationResult {
	if conv == nil {
		return ValidationResult{Valid: true, Message: message}
	}

	result := ValidationResult{Message: message}

	// Extract first line (header) for validation.
	header := strings.SplitN(message, "\n", 2)[0]
	header = strings.TrimSpace(header)

	if header == "" {
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationRequired,
			Field:    "header",
			Expected: "non-empty commit message",
			Actual:   "",
		})
		result.Valid = false
		return result
	}

	// Check max length.
	if conv.MaxLength > 0 && len(header) > conv.MaxLength {
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationMaxLength,
			Field:    "header",
			Expected: fmt.Sprintf("max %d characters", conv.MaxLength),
			Actual:   fmt.Sprintf("%d characters", len(header)),
		})
	}

	// Check pattern match.
	if !conv.Pattern.MatchString(header) {
		result.Violations = append(result.Violations, Violation{
			Type:       ViolationPattern,
			Field:      "header",
			Expected:   conv.Pattern.String(),
			Actual:     header,
			Suggestion: suggestFix(header, conv),
		})
	} else {
		// Pattern matches; check semantic rules.
		validateSemantics(header, conv, &result)
	}

	result.Valid = len(result.Violations) == 0
	return result
}

// validateSemantics checks type and scope against allowed lists.
func validateSemantics(header string, conv *Convention, result *ValidationResult) {
	commitType := extractType(header)

	// Check type validity.
	if len(conv.Types) > 0 && commitType != "" {
		found := slices.Contains(conv.Types, commitType)
		if !found {
			result.Violations = append(result.Violations, Violation{
				Type:     ViolationInvalidType,
				Field:    "type",
				Expected: strings.Join(conv.Types, ", "),
				Actual:   commitType,
			})
		}
	}

	// Check scope validity (only when scopes are defined).
	scope := extractScope(header, conv.ScopeDelimiter)
	if len(conv.Scopes) > 0 && scope != "" {
		found := slices.Contains(conv.Scopes, scope)
		if !found {
			result.Violations = append(result.Violations, Violation{
				Type:     ViolationInvalidScope,
				Field:    "scope",
				Expected: strings.Join(conv.Scopes, ", "),
				Actual:   scope,
			})
		}
	}
}

// extractType extracts the commit type from the header.
// e.g., "feat(auth): add JWT" -> "feat"
func extractType(header string) string {
	for i, c := range header {
		if c == '(' || c == ':' || c == '!' {
			return header[:i]
		}
	}
	return ""
}

// extractScope는 헤더에서 scope를 추출한다.
// delim이 "[]"이면 "type!?: " 직후 위치에서만 대괄호 scope를 추출하고,
// "()"이면 type 직후 소괄호에서 추출한다.
// 빈 문자열이면 기본값 "()" 사용 (하위호환).
func extractScope(header string, delim string) string {
	if delim == "" {
		delim = "()"
	}

	// bracket-scope: "type!?: [Scope] desc" — ": " 직후에서만 추출
	if delim == "[]" {
		colonSpace := strings.Index(header, ": ")
		if colonSpace < 0 {
			return ""
		}
		afterColon := colonSpace + 2
		if afterColon >= len(header) || header[afterColon] != '[' {
			return ""
		}
		end := strings.IndexByte(header[afterColon:], ']')
		if end < 0 {
			return ""
		}
		return header[afterColon+1 : afterColon+end]
	}

	// conventional: "type(scope): desc" — 첫 번째 소괄호에서 추출
	openByte := delim[0]
	closeByte := delim[1]
	start := strings.IndexByte(header, openByte)
	if start < 0 {
		return ""
	}
	end := strings.IndexByte(header[start:], closeByte)
	if end < 0 {
		return ""
	}
	return header[start+1 : start+end]
}

// suggestFix는 유효한 커밋 메시지 형식을 제안한다.
func suggestFix(header string, conv *Convention) string {
	lower := strings.ToLower(header)

	suggestedType := "chore"
	switch {
	case strings.Contains(lower, "fix") || strings.Contains(lower, "bug"):
		suggestedType = "fix"
	case strings.Contains(lower, "add") || strings.Contains(lower, "feat") || strings.Contains(lower, "new"):
		suggestedType = "feat"
	case strings.Contains(lower, "doc") || strings.Contains(lower, "readme"):
		suggestedType = "docs"
	case strings.Contains(lower, "test"):
		suggestedType = "test"
	case strings.Contains(lower, "refactor") || strings.Contains(lower, "clean"):
		suggestedType = "refactor"
	}

	desc := strings.TrimSpace(header)
	if len(desc) > 0 {
		// 첫 글자 소문자 변환
		desc = strings.ToLower(desc[:1]) + desc[1:]
	}

	// bracket-scope 컨벤션이면 [Scope] 형식으로 제안
	if conv != nil && conv.ScopeDelimiter == "[]" {
		return suggestedType + ": [Scope] " + desc
	}

	return suggestedType + ": " + desc
}
