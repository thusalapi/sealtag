package cli

import (
	"encoding/json"
	"io"
)

// writeJSON prints v as indented JSON. Every command supports --json so CI and other
// tools can read sealtag's output without parsing tables.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
