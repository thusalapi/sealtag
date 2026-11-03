package npm

import (
	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/textedit"
)

// SetVersion updates "version" in package.json, keeping its formatting.
// package-lock.json also stores the version, so it is updated when present.
func (Ecosystem) SetVersion(root string, p ecosystem.Package, version string) ([]string, error) {
	set := func(b []byte) ([]byte, error) { return textedit.SetJSONTopLevelString(b, "version", version) }
	var changed []string
	if ok, err := ecosystem.EditFile(root, p.Manifest, set); err != nil {
		return nil, err
	} else if ok {
		changed = append(changed, p.Manifest)
	}
	lock := joinRel(p.Dir, "package-lock.json")
	if ok, err := ecosystem.EditFile(root, lock, set); err == nil && ok {
		changed = append(changed, lock)
	}
	return changed, nil
}

func joinRel(dir, name string) string {
	if dir == "." || dir == "" {
		return name
	}
	return dir + "/" + name
}
