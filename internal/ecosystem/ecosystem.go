// Package ecosystem defines the shared types every registry implementation uses.
package ecosystem

// Package is one publishable unit found in a repository.
type Package struct {
	Ecosystem string `json:"ecosystem"` // "npm", "pypi", "cargo" or "go"
	Name      string `json:"name"`
	Version   string `json:"version"`        // empty when not stored in the manifest
	Dir       string `json:"dir"`            // directory that holds the manifest
	Manifest  string `json:"manifest"`       // path to the manifest file
	Private   bool   `json:"private"`        // true when it must not be published
	Note      string `json:"note,omitempty"` // detail, for example "version is dynamic"
}

// Ecosystem is implemented once per registry. See docs/decisions/0002.
type Ecosystem interface {
	// Name returns the short ecosystem id, such as "npm".
	Name() string
	// Detect returns the package defined directly in dir, or nil if there is none.
	Detect(dir string) (*Package, error)
}
