// Package ecosystem defines the shared types every registry implementation uses.
package ecosystem

// Package is one publishable unit found in a repository.
type Package struct {
	Ecosystem string // "npm", "pypi", "cargo" or "go"
	Name      string
	Version   string // empty when the version is not stored in the manifest
	Dir       string // directory that holds the manifest
	Manifest  string // path to the manifest file
	Private   bool   // true when the manifest says the package must not be published
	Note      string // human-readable detail, for example "version is dynamic"
}

// Ecosystem is implemented once per registry. See docs/decisions/0002.
type Ecosystem interface {
	// Name returns the short ecosystem id, such as "npm".
	Name() string
	// Detect returns the package defined directly in dir, or nil if there is none.
	Detect(dir string) (*Package, error)
}
