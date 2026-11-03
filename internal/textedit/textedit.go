// Package textedit changes one value in a manifest file without reformatting the
// rest of it. Parsing and re-encoding would lose comments, key order and spacing,
// which makes release diffs noisy and annoys maintainers.
package textedit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ErrNotFound means the key was not present where expected.
var ErrNotFound = errors.New("key not found")

// SetJSONTopLevelString replaces the string value of a top-level key in a JSON
// object. Nested keys with the same name (for example inside "engines") are ignored.
func SetJSONTopLevelString(data []byte, key, value string) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	depth := 0
	expectKey := false
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%q: %w", key, ErrNotFound)
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{', '[':
				depth++
				expectKey = t == '{' && depth == 1
			case '}', ']':
				depth--
				expectKey = depth == 1
			}
			continue
		case string:
			if depth == 1 && expectKey && t == key {
				vStart := dec.InputOffset()
				vTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				if _, ok := vTok.(string); !ok {
					return nil, fmt.Errorf("%q is not a string", key)
				}
				vEnd := dec.InputOffset()
				// vStart sits before the ':' and any spaces. Find the opening quote.
				q := bytes.IndexByte(data[vStart:vEnd], '"')
				enc, _ := json.Marshal(value)
				out := append([]byte{}, data[:int(vStart)+q]...)
				out = append(out, enc...)
				return append(out, data[vEnd:]...), nil
			}
		}
		// At depth 1 keys and values alternate.
		if depth == 1 {
			expectKey = !expectKey
		}
	}
}

var tableHeader = regexp.MustCompile(`^\s*\[([^\[\]]+)\]\s*(#.*)?$`)

// SetTOMLString replaces `key = "..."` inside [table]. table "" means the top level.
// Only simple string values on one line are supported, which covers the version
// fields sealtag edits.
func SetTOMLString(data []byte, table, key, value string) ([]byte, error) {
	keyLine := regexp.MustCompile(`^(\s*` + regexp.QuoteMeta(key) + `\s*=\s*)(["'])([^"']*)(["'])(.*)$`)
	lines := strings.SplitAfter(string(data), "\n")
	current := ""
	for i, line := range lines {
		body := strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(strings.TrimSpace(body), "[[") {
			current = "[[array]]" // array tables are never the target
			continue
		}
		if m := tableHeader.FindStringSubmatch(body); m != nil {
			current = strings.TrimSpace(m[1])
			continue
		}
		if current != table {
			continue
		}
		if m := keyLine.FindStringSubmatch(body); m != nil {
			eol := line[len(body):]
			lines[i] = m[1] + m[2] + value + m[4] + m[5] + eol
			return []byte(strings.Join(lines, "")), nil
		}
	}
	return nil, fmt.Errorf("[%s] %s: %w", table, key, ErrNotFound)
}

// SetPythonVersion replaces `__version__ = "..."` in a Python source file.
func SetPythonVersion(data []byte, value string) ([]byte, error) {
	re := regexp.MustCompile(`(?m)^(__version__\s*=\s*)(["'])([^"']*)(["'])`)
	loc := re.FindSubmatchIndex(data)
	if loc == nil {
		return nil, fmt.Errorf("__version__: %w", ErrNotFound)
	}
	out := append([]byte{}, data[:loc[6]]...)
	out = append(out, value...)
	return append(out, data[loc[7]:]...), nil
}

// PythonVersion reads `__version__ = "..."` from a Python source file.
func PythonVersion(data []byte) (string, bool) {
	m := regexp.MustCompile(`(?m)^__version__\s*=\s*["']([^"']*)["']`).FindSubmatch(data)
	if m == nil {
		return "", false
	}
	return string(m[1]), true
}
