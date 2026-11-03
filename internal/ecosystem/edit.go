package ecosystem

import (
	"os"
	"path/filepath"
)

// EditFile applies fn to the file at root/rel and writes it back with the same
// permissions. It is a no-op when fn returns the same bytes.
func EditFile(root, rel string, fn func([]byte) ([]byte, error)) (changed bool, err error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := fn(data)
	if err != nil {
		return false, err
	}
	if string(out) == string(data) {
		return false, nil
	}
	return true, os.WriteFile(path, out, info.Mode().Perm())
}
