package llm

import (
	"errors"
	"strings"
)

// ErrNoJSON is returned when no JSON object can be located in model output.
var ErrNoJSON = errors.New("llm: no JSON object found in output")

// ExtractJSON pulls the outermost JSON object from model output that may be
// wrapped in prose or ```json fences. It does not validate the object's shape;
// callers unmarshal + validate separately. Returns ErrNoJSON if none is found.
func ExtractJSON(s string) (string, error) {
	s = strings.TrimSpace(s)
	// Strip code fences if present.
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			// drop an optional language tag line (e.g. "json")
			if strings.TrimSpace(s[:i]) != "" && !strings.ContainsAny(s[:i], "{}") {
				s = s[i+1:]
			}
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end < 0 || end < start {
		return "", ErrNoJSON
	}
	return s[start : end+1], nil
}
