// Package fileset lists the files in a directory that belong to the project.
//
// Inside a git repository it asks git, so .gitignore rules are respected exactly as
// git applies them. Outside git it walks the directory and skips well-known
// dependency and build folders.
package fileset

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// List returns paths relative to root, using forward slashes, sorted.
// skipDir is used only by the fallback walk.
func List(root string, skipDir func(string) bool) ([]string, error) {
	if files, ok := GitFiles(root); ok {
		return filterSkipped(files, skipDir), nil
	}
	return walk(root, skipDir)
}

// GitFiles lists tracked files plus untracked files that are not ignored, relative
// to root. ok is false when root is not inside a git repository.
// Untracked files matter: a new .env that was never committed can still be packed.
func GitFiles(root string) ([]string, bool) {
	cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	seen := map[string]bool{}
	var files []string
	for _, f := range strings.Split(string(out), "\x00") {
		if f == "" || seen[f] {
			continue
		}
		// ls-files also lists files deleted from the working tree. Keep only real files.
		if !exists(filepath.Join(root, filepath.FromSlash(f))) {
			continue
		}
		seen[f] = true
		files = append(files, f)
	}
	sort.Strings(files)
	return files, true
}

// filterSkipped still drops skipped folders such as testdata, even when tracked.
func filterSkipped(files []string, skipDir func(string) bool) []string {
	if skipDir == nil {
		return files
	}
	out := files[:0]
outer:
	for _, f := range files {
		parts := strings.Split(f, "/")
		for _, dir := range parts[:len(parts)-1] {
			if skipDir(dir) {
				continue outer
			}
		}
		out = append(out, f)
	}
	return out
}

func walk(root string, skipDir func(string) bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDir != nil && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
