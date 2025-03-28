package fs

import (
	"errors"
	"path/filepath"
	"strconv"

	"github.com/transientvariable/schema-go"
)

// FileMetadata converts a file system entry and produces a schema.File.
func FileMetadata(fsys FS, entry *Entry) (*schema.File, error) {
	if fsys == nil {
		return nil, errors.New("fs: file system is required")
	}

	if entry == nil {
		return nil, errors.New("fs: entry is required")
	}

	r, err := fsys.Root()
	if err != nil {
		return nil, err
	}

	mtime := entry.ModTime()
	m := &schema.File{
		Ctime:     &mtime,
		GID:       itoa(int(entry.Attributes().GID())),
		Directory: filepath.Join(fsys.PathSeparator(), r, filepath.Dir(entry.Path())),
		Inode:     itoa(int(entry.Attributes().Inode())),
		Mode:      itoa(int(entry.Mode())),
		Mtime:     &mtime,
		Name:      entry.Name(),
		Owner:     entry.Attributes().Owner(),
		Path:      filepath.Join(fsys.PathSeparator(), r, entry.Path()),
		UID:       itoa(int(entry.Attributes().UID())),
	}

	if !entry.IsDir() {
		m.MimeType = entry.Attributes().MimeType()
		m.Size = entry.Size()
	}
	return m, nil
}

func itoa(v int) string {
	if v > 0 {
		return strconv.Itoa(v)
	}
	return ""
}
