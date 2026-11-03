package pypi

import (
	"errors"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/textedit"
)

// SetVersion writes the version where the project keeps it: a __version__ file for
// hatch dynamic versions, [tool.poetry] for Poetry, or [project] otherwise.
func (Ecosystem) SetVersion(root string, p ecosystem.Package, version string) ([]string, error) {
	file, edit := p.Manifest, func(b []byte) ([]byte, error) {
		return textedit.SetTOMLString(b, "project", "version", version)
	}
	switch {
	case p.VersionFile != "":
		file = p.VersionFile
		edit = func(b []byte) ([]byte, error) { return textedit.SetPythonVersion(b, version) }
	case p.Note == "poetry":
		edit = func(b []byte) ([]byte, error) { return textedit.SetTOMLString(b, "tool.poetry", "version", version) }
	case p.Version == "":
		return nil, errors.New("version is dynamic and sealtag does not know where it is stored")
	}
	ok, err := ecosystem.EditFile(root, file, edit)
	if err != nil || !ok {
		return nil, err
	}
	return []string{file}, nil
}
