package fs

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	gofs "io/fs"
)

// CleanPath cleans the path p returns a lexically valid path.
func CleanPath(fsys FS, p string) (string, error) {
	if fsys == nil {
		return p, errors.New("file system is required")
	}

	if p = strings.TrimSpace(p); !gofs.ValidPath(p) {
		return p, fmt.Errorf("%s: %w", p, gofs.ErrInvalid)
	}

	if strings.HasSuffix(p, fsys.PathSeparator()) {
		p = p[:len(p)-1]
	}

	if vol := filepath.VolumeName(p); len(vol) > 0 {
		p = p[len(vol)-1:]
	}
	return p, nil
}

// SplitPath splits a path using the path separator from the provided file system.
//
// The returned slice will have empty substrings removed.
func SplitPath(fsys FS, p string) ([]string, error) {
	path, err := CleanPath(fsys, p)
	if err != nil {
		return nil, err
	}

	var e []string
	for _, s := range strings.Split(path, fsys.PathSeparator()) {
		if s != "" {
			e = append(e, s)
		}
	}
	return e, nil
}

// EndsWithDot reports whether the final component of the path is ".".
func EndsWithDot(fsys FS, path string) bool {
	if path == "." {
		return true
	}

	if len(path) >= 2 && path[len(path)-1] == '.' && fsys.PathSeparator() == string(path[len(path)-2]) {
		return true
	}
	return false
}
