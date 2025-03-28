package memfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/transientvariable/collection-go"
	"github.com/transientvariable/collection-go/trie"
	"github.com/transientvariable/fs-go"
	"github.com/transientvariable/log-go"
	"github.com/transientvariable/support-go"

	gofs "io/fs"
)

const (
	pathSeparator = string(os.PathSeparator)
	modePerm      = 0664
)

var _ fs.FS = (*MemFS)(nil)

// MemFS in-memory file system provider that implements fs.FS.
//
// Unless otherwise specified, all operations are transient and will be lost when the runtime exits.
type MemFS struct {
	closed  bool
	entry   *fs.Entry
	entries trie.Trie
	mutex   sync.Mutex
}

// New creates a new MemFS.
func New() (*MemFS, error) {
	return newDir(pathSeparator, modePerm, fs.WithPathValidator(func(p string) bool { return true }))
}

// Close ...
func (m *MemFS) Close() error {
	if m == nil {
		return gofs.ErrInvalid
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.entry.Name() != pathSeparator {
		return nil
	}

	if !m.closed {
		m.closed = true
		return nil
	}
	return fmt.Errorf("memfs: %w", gofs.ErrClosed)
}

// Create ...
func (m *MemFS) Create(name string) (fs.File, error) {
	log.Debug("[memfs] create", log.String("name", name))
	return m.open("create", name, fs.O_RDWR|fs.O_CREATE|fs.O_TRUNC, modePerm)
}

// Glob ...
func (m *MemFS) Glob(pattern string) ([]string, error) {
	log.Debug("[memfs] glob", log.String("pattern", pattern))

	var matches []string
	err := gofs.WalkDir(m, ".", func(path string, entry gofs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		if !errors.Is(err, &gofs.PathError{}) {
			return matches, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "glob", Err: err})
		}
		return matches, err
	}
	return matches, nil
}

// Mkdir ...
func (m *MemFS) Mkdir(name string, perm gofs.FileMode) error {
	log.Debug("[memfs] mkdir", log.String("name", name))

	name, err := fs.CleanPath(m, name)
	if err != nil {
		return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "mkdir", Path: name, Err: err})
	}

	if _, err := m.Stat(name); err != nil {
		if !errors.Is(err, gofs.ErrNotExist) {
			return fmt.Errorf("memfs: %w", err)
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, err := mkdir(m, name, perm); err != nil {
		return fmt.Errorf("memfs: %w", err)
	}
	return nil
}

// MkdirAll ...
func (m *MemFS) MkdirAll(path string, mode gofs.FileMode) error {
	log.Debug("[memfs] mkdirAll", log.String("path", path), log.String("mode", mode.String()))

	path, err := fs.CleanPath(m, path)
	if err != nil {
		return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "mkdirAll", Path: path, Err: err})
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, err := mkdirAll(m, path, mode); err != nil {
		return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "mkdirAll", Path: path, Err: err})
	}
	return nil
}

// Open opens the named File.
func (m *MemFS) Open(name string) (gofs.File, error) {
	log.Debug("[memfs] open", log.String("name", name))
	return m.open("open", name, fs.O_RDONLY, 0)
}

// OpenFile ...
func (m *MemFS) OpenFile(name string, flag int, mode gofs.FileMode) (fs.File, error) {
	log.Debug("[memfs] openFile", log.String("name", name), log.Int("flag", flag), log.String("mode", mode.String()))
	return m.open("openFile", name, flag, mode)
}

// PathSeparator ...
func (m *MemFS) PathSeparator() string {
	return pathSeparator
}

// Provider ...
func (m *MemFS) Provider() string {
	return "memfs"
}

// ReadDir ...
func (m *MemFS) ReadDir(name string) ([]gofs.DirEntry, error) {
	log.Debug("[memfs] readDir", log.String("name", name))

	sub, err := sub(m, name)
	if err != nil {
		return nil, err
	}

	mfs := sub.(*MemFS)
	de, err := newDirIterator(mfs).NextN(-1)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "readDir", Path: mfs.entry.Path(), Err: err})
	}

	entries := make([]gofs.DirEntry, len(de))
	for i, e := range de {
		entries[i] = e
	}
	return entries, nil
}

// ReadFile ...
func (m *MemFS) ReadFile(name string) ([]byte, error) {
	log.Debug("[memfs] readFile", log.String("name", name))

	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	defer func(f gofs.File) {
		if err := f.Close(); err != nil {
			log.Error("[memfs] readFile", log.Err(err))
		}
	}(f)

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "readFile", Path: name, Err: err})
	}
	return b, nil
}

// Remove ...
func (m *MemFS) Remove(name string) error {
	log.Debug("[memfs] remove", log.String("name", name))
	return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "remove", Path: name, Err: errors.New("not implemented")})
}

// RemoveAll ...
func (m *MemFS) RemoveAll(path string) error {
	log.Debug("[memfs] removeAll", log.String("path", path))
	return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "removeAll", Path: path, Err: errors.New("not implemented")})
}

// Rename ...
func (m *MemFS) Rename(oldpath string, newpath string) error {
	log.Debug("[memfs] rename", log.String("old_path", oldpath), log.String("new_path", newpath))
	return fmt.Errorf("memfs: %w", &gofs.PathError{Op: "rename", Err: errors.New("not implemented")})
}

// Root ...
func (m *MemFS) Root() (string, error) {
	return pathSeparator, nil
}

// Stat ...
func (m *MemFS) Stat(name string) (gofs.FileInfo, error) {
	log.Debug("[memfs] stat", log.String("name", name))

	e, err := stat(m, name)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "stat", Path: name, Err: err})
	}
	return e.Stat()
}

// Sub ...
func (m *MemFS) Sub(dir string) (gofs.FS, error) {
	log.Debug("[memfs] sub", log.String("current", m.entry.Name()), log.String("dir", dir))

	sub, err := sub(m, dir)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "sub", Path: dir, Err: err})
	}
	return sub, nil
}

// WriteFile ...
func (m *MemFS) WriteFile(name string, data []byte, mode gofs.FileMode) error {
	log.Debug("[memfs] writeFile",
		log.String("name", name),
		log.Int("content_length", len(data)),
		log.String("mode", mode.String()),
	)

	f, err := m.open("writeFile", name, fs.O_RDWR|fs.O_CREATE|fs.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func(f *File) {
		if err := f.Close(); err != nil {
			log.Error("[memfs] writeFile", log.Err(err))
		}
	}(f)

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

// String returns a string representation of MemFS.
func (m *MemFS) String() string {
	s := make(map[string]any)
	s["mode"] = m.entry.Mode().String()
	s["mod_time"] = m.entry.ModTime()
	s["Name"] = m.entry.Name()

	entries, err := list(m)
	if err != nil {
		entries = append(entries, err.Error())
	}
	s["list"] = entries
	return string(support.ToJSONFormatted(s))
}

func (m *MemFS) open(op string, name string, flag int, mode gofs.FileMode) (*File, error) {
	name, err := fs.CleanPath(m, name)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
	}

	s, err := stat(m, name)
	if err != nil {
		if errors.Is(err, gofs.ErrNotExist) && flag&fs.O_CREATE != 0 {
			return create(m, name, flag, mode)
		}
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
	}

	if s != nil {
		switch s.Data().(type) {
		case *fd:
			fd := s.Data().(*fd)
			if !fd.entry.IsDir() {
				return newFile(fd, flag)
			}
			return newFile(fd, fs.O_RDONLY)
		case *MemFS:
			mfs := s.Data().(*MemFS)
			fd, err := newfd(mfs, ".", fs.O_RDONLY, mfs.entry.Mode())
			if err != nil {
				return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
			}
			return newFile(fd, fs.O_RDONLY)
		default:
			return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: gofs.ErrInvalid})
		}
	}

	p, err := fs.SplitPath(m, name)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
	}

	if len(p) > 1 {
		e, err := stat(m, filepath.Dir(name))
		if err != nil {
			return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
		}

		fd, err := newfd(e.Data().(*MemFS), filepath.Base(name), flag, mode)
		if err != nil {
			return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
		}
		return newFile(fd, flag)
	}

	fd, err := newfd(m, name, flag, mode)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: op, Path: name, Err: err})
	}
	return newFile(fd, flag)
}

func create(mfs *MemFS, name string, flag int, mode gofs.FileMode) (*File, error) {
	mfs.mutex.Lock()
	defer mfs.mutex.Unlock()

	if mode&gofs.ModeDir != 0 {
		log.Trace("[memfs:create] directory mode bits set, creating path as directory", log.String("name", name))

		dir, err := mkdirAll(mfs, name, mode)
		if err != nil {
			return nil, err
		}

		fd, err := newfd(dir, ".", flag, mode)
		if err != nil {
			return nil, err
		}

		file, err := newFile(fd, flag)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	p, err := fs.SplitPath(mfs, name)
	if err != nil {
		return nil, err
	}

	if len(p) == 1 {
		fd, err := newfd(mfs, name, flag, mode)
		if err != nil {
			return nil, err
		}
		return newFile(fd, flag)
	}

	log.Trace("[memfs:create] creating directory for file", log.String("directory", filepath.Dir(name)))

	dir, err := mkdirAll(mfs, filepath.Dir(name), mode)
	if err != nil {
		return nil, err
	}

	log.Trace("[memfs:create]", log.String("directory", dir.entry.Name()), log.String("name", filepath.Base(name)))

	fd, err := newfd(dir, filepath.Base(name), flag, mode)
	if err != nil {
		return nil, err
	}
	return newFile(fd, flag)
}

func entry(mfs *MemFS, name string) (*fsEntry, error) {
	e, err := mfs.entries.Entry(name)
	if err != nil {
		if errors.Is(err, collection.ErrCollectionEmpty) || errors.Is(err, collection.ErrNotFound) {
			return nil, gofs.ErrNotExist
		}
		return nil, err
	}

	fse, ok := e.(*fsEntry)
	if !ok {
		return nil, gofs.ErrInvalid
	}
	return fse, nil
}

func find(mfs *MemFS, name string) (*fsEntry, error) {
	if name == "." {
		return entry(mfs, name)
	}

	n, err := fs.SplitPath(mfs, name)
	if err != nil {
		return nil, err
	}

	if len(n) > 1 {
		e, err := entry(mfs, n[0])
		if err != nil {
			return nil, err
		}

		if e.entry.IsDir() {
			return find(e.Data().(*MemFS), strings.Join(n[1:], pathSeparator))
		}
		return nil, gofs.ErrNotExist
	}
	return entry(mfs, name)
}

func list(mfs *MemFS) ([]string, error) {
	var entries []string
	err := gofs.WalkDir(mfs, ".", func(path string, entry gofs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fi, err := entry.Info()
		if err != nil {
			return err
		}

		entries = append(entries, fmt.Sprintf("%s: size: %d, mode: %s, mode_type: %s", path, fi.Size(), fi.Mode(), fi.Mode().Type()))
		return nil
	})
	if err != nil {
		return entries, err
	}
	return entries, nil
}

func mkdir(mfs *MemFS, name string, mode gofs.FileMode) (*MemFS, error) {
	if name == "." {
		return nil, &gofs.PathError{Op: "mkdir", Path: name, Err: gofs.ErrInvalid}
	}

	p, err := fs.SplitPath(mfs, name)
	if err != nil {
		return nil, err
	}

	if len(p) > 1 {
		dir := filepath.Dir(name)
		e, err := stat(mfs, dir)
		if err != nil {
			return nil, &gofs.PathError{Op: "mkdir", Path: dir, Err: gofs.ErrInvalid}
		}
		mfs = e.Data().(*MemFS)
	}

	if !mfs.entry.IsDir() {
		return mfs, &gofs.PathError{Op: "mkdir", Path: filepath.Dir(name), Err: fs.ErrNotDir}
	}

	// TODO: Check writable permission of parent?

	if _, err := entry(mfs, filepath.Base(name)); err != nil {
		if errors.Is(err, gofs.ErrNotExist) {
			n, err := newDir(filepath.Base(name), mode)
			if err != nil {
				return nil, &gofs.PathError{Op: "mkdir", Path: name, Err: err}
			}

			if err = mfs.entries.AddEntry(&fsEntry{
				entry: n.entry,
				data:  n,
			}); err != nil {
				return nil, &gofs.PathError{Op: "mkdir", Path: name, Err: err}
			}

			if err := mfs.entry.SetModTime(time.Now()); err != nil {
				return nil, &gofs.PathError{Op: "mkdir", Path: name, Err: err}
			}
			return n, nil
		}
		return nil, &gofs.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return mfs, &gofs.PathError{Op: "mkdir", Path: name, Err: gofs.ErrExist}
}

func mkdirAll(mfs *MemFS, path string, mode gofs.FileMode) (*MemFS, error) {
	p, err := fs.SplitPath(mfs, path)
	if err != nil {
		return nil, err
	}

	for _, dir := range p {
		s, err := stat(mfs, dir)
		if err != nil {
			if !errors.Is(err, gofs.ErrNotExist) {
				return nil, err
			}
		}

		if s != nil {
			mfs = s.Data().(*MemFS)
			continue
		}

		d, err := mkdir(mfs, dir, mode)
		if err != nil {
			return nil, err
		}
		mfs = d
	}
	return mfs, nil
}

func newDir(name string, mode gofs.FileMode, entryOptions ...func(*fs.Entry)) (*MemFS, error) {
	attrs, err := fs.NewAttributes(fs.WithMode(uint32(mode | gofs.ModeDir)))
	if err != nil {
		return nil, err
	}

	dir, err := fs.NewEntry(name, append(entryOptions, fs.WithAttributes(attrs))...)
	if err != nil {
		return nil, err
	}

	entries, err := trie.New()
	if err != nil {
		return nil, err
	}

	mfs := &MemFS{entry: dir, entries: entries}
	_, err = newfd(mfs, ".", fs.O_CREATE, dir.Mode())
	if err != nil {
		return nil, err
	}
	return mfs, nil
}

func stat(mfs *MemFS, name string) (*fsEntry, error) {
	name, err := fs.CleanPath(mfs, name)
	if err != nil {
		return nil, err
	}

	e, err := find(mfs, name)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func sub(mfs *MemFS, dir string) (gofs.SubFS, error) {
	dir, err := fs.CleanPath(mfs, dir)
	if err != nil {
		return nil, err
	}

	if dir == "." {
		return mfs, nil
	}

	e, err := find(mfs, dir)
	if err != nil {
		return nil, fmt.Errorf("memfs: %w", &gofs.PathError{Op: "sub", Path: dir, Err: err})
	}
	return e.Data().(*MemFS), nil
}
