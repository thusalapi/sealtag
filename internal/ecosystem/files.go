package ecosystem

import "errors"

// File is one file that will be inside a published artifact.
type File struct {
	Path string `json:"path"` // relative to the package directory, forward slashes
	Size int64  `json:"size"`
	// Generated is true for files the registry tool writes into the archive itself,
	// such as Cargo.toml.orig. They are not on disk, so they cannot be scanned.
	Generated bool `json:"generated,omitempty"`
}

// Packer is implemented by ecosystems that can say exactly which files a release
// would contain. It is a separate interface so ecosystems can add it one at a time.
type Packer interface {
	// ShippedFiles lists the files that publishing the package in dir would upload.
	ShippedFiles(dir string) ([]File, error)
}

// ErrToolMissing means the registry's own tool (npm, cargo, ...) is not installed.
var ErrToolMissing = errors.New("required tool is not installed")

// ErrNotSupported means sealtag cannot list files for this ecosystem yet.
var ErrNotSupported = errors.New("file listing not supported for this ecosystem yet")
