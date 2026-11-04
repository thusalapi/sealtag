package ecosystem

import (
	"context"
	"errors"
	"io"

	"github.com/thusalapi/sealtag/internal/oidc"
)

// File is one file that will be inside a published artifact.
type File struct {
	Path string `json:"path"` // relative to the package directory, forward slashes
	Size int64  `json:"size"`
	// Generated is true for files the registry tool writes into the archive itself,
	// such as Cargo.toml.orig. They are not on disk, so they cannot be scanned.
	Generated bool `json:"generated,omitempty"`
	// Artifact names the built archive the file came from (for example a wheel),
	// when the ecosystem had to build the package to know its contents.
	Artifact string `json:"artifact,omitempty"`
	// Data holds the file content read from the artifact, so the scan checks the
	// bytes that really ship. Nil when the file is read from disk instead.
	Data []byte `json:"-"`
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

// Versioner is implemented by ecosystems that store the version in a file.
type Versioner interface {
	// SetVersion writes version for the package and returns the files it changed,
	// relative to the package directory's root (the repository root).
	SetVersion(root string, p Package, version string) ([]string, error)
}

// PublishOptions are passed to every Publisher.
type PublishOptions struct {
	DryRun bool      // do everything except contacting a registry
	Log    io.Writer // progress messages and GitHub Actions masking commands
	OIDC   oidc.Client
}

// Publisher uploads a package with trusted publishing.
type Publisher interface {
	Publish(ctx context.Context, dir string, p Package, opts PublishOptions) error
}

// Verifier installs a package from its built artifact into an empty, temporary
// environment and checks that it loads. This catches a package that works in the
// repository but breaks once published, for example because a needed file is
// excluded or a generated file is missing.
type Verifier interface {
	Verify(ctx context.Context, dir string, p Package, opts VerifyOptions) error
}

// VerifyOptions control where the clean-room install runs.
type VerifyOptions struct {
	Log io.Writer
	// Container runs the install and load steps in a fresh container instead of a
	// temporary folder on the host. A container also hides the host's caches and
	// globally installed packages, so "it only works on my machine" cannot pass.
	Container bool
}
