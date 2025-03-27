package fs

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/transientvariable/collection"
	"github.com/transientvariable/log"

	gofs "io/fs"
)

var (
	defaultFS FS
	mutex     sync.Mutex
	once      sync.Once
)

// Initialize the default file system provider (e.g. memfs.MemFS).
func init() {
	once.Do(func() {
		// Use osfs as opposed provider from config to ensure we have a working file system.
		fsys, err := New()
		if err != nil {
			panic(err)
		}

		if err := SetDefault(fsys); err != nil {
			panic(err)
		}
	})
}

const (
	O_RDONLY = os.O_RDONLY
	O_WRONLY = os.O_WRONLY
	O_RDWR   = os.O_RDWR
	O_APPEND = os.O_APPEND
	O_CREATE = os.O_CREATE
	O_TRUNC  = os.O_TRUNC

	// MaxContentLen defines the maximum size in bytes for a File.
	MaxContentLen = int(^uint(0) >> 1)
)

// DirIterator defines the behavior for iterating over entries in a directory.
type DirIterator interface {
	collection.Iterator[*Entry]

	// NextN returns a slice containing the next n directory list. Dot list "." are skipped.
	//
	// The error io.EOF is returned if there are no remaining list left to iterate.
	NextN(n int) ([]*Entry, error)
}

// File defines the behavior for providing access to a single file. This interface is an extension of the fs.Name
// interface and defines additional behavior for read/write operations.
type File interface {
	gofs.File
	gofs.ReadDirFile
	io.ReaderAt
	io.ReaderFrom
	io.Seeker
	io.Writer
}

// Readable defines the behavior for providing read access to a hierarchical file system.
type Readable interface {
	gofs.FS
	gofs.GlobFS
	gofs.ReadFileFS
	gofs.ReadDirFS
	gofs.StatFS
	gofs.SubFS
}

// Writable defines the behavior for providing write access to a hierarchical file system.
type Writable interface {
	// Create ...
	Create(name string) (File, error)

	// Mkdir ...
	Mkdir(name string, perm gofs.FileMode) error

	// MkdirAll ...
	MkdirAll(path string, perm gofs.FileMode) error

	// OpenFile ...
	OpenFile(name string, flag int, perm gofs.FileMode) (File, error)

	// Remove ...
	Remove(name string) error

	// RemoveAll ...
	RemoveAll(path string) error

	// Rename ...
	Rename(oldpath string, newpath string) error

	// WriteFile ...
	WriteFile(name string, data []byte, perm gofs.FileMode) error
}

// FS defines the basic behavior for providing access to a hierarchical file system.
type FS interface {
	Readable
	Writable

	// PathSeparator ...
	PathSeparator() string

	// Provider ...
	Provider() string

	// Root ...
	Root() (string, error)

	// Close ...
	Close() error
}

// SetDefault sets the default file system backend.
func SetDefault(fs FS) error {
	if fs == nil {
		return errors.New("fs: file system is required")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if defaultFS != nil {
		log.Info("[fs] setting default file system", log.String("provider", fs.Provider()))
	}
	defaultFS = fs
	return nil
}

// Default returns the current default for the file system backend.
func Default() FS {
	return defaultFS
}

// Create ...
func Create(name string) (File, error) {
	return Default().Create(name)
}

// Glob ...
func Glob(pattern string) ([]string, error) {
	return Default().Glob(pattern)
}

// Mkdir ...
func Mkdir(name string, perm gofs.FileMode) error {
	return Default().Mkdir(name, perm)
}

// MkdirAll ...
func MkdirAll(path string, perm gofs.FileMode) error {
	return Default().MkdirAll(path, perm)
}

// Open ...
func Open(name string) (gofs.File, error) {
	return Default().Open(name)
}

// OpenFile ...
func OpenFile(name string, flag int, perm gofs.FileMode) (File, error) {
	return Default().OpenFile(name, flag, perm)
}

// ReadDir ...
func ReadDir(name string) ([]gofs.DirEntry, error) {
	return Default().ReadDir(name)
}

// ReadFile ...
func ReadFile(name string) ([]byte, error) {
	return Default().ReadFile(name)
}

// Remove ...
func Remove(name string) error {
	return Default().Remove(name)
}

// RemoveAll ...
func RemoveAll(path string) error {
	return Default().RemoveAll(path)
}

// Rename ...
func Rename(oldpath string, newpath string) error {
	return Default().Rename(oldpath, newpath)
}

// Root ...
func Root() (string, error) {
	return Default().Root()
}

// Stat ...
func Stat(name string) (gofs.FileInfo, error) {
	return Default().Stat(name)
}

// Sub ...
func Sub(dir string) (gofs.FS, error) {
	return Default().Sub(dir)
}

// WriteFile ...
func WriteFile(name string, data []byte, perm gofs.FileMode) error {
	return Default().WriteFile(name, data, perm)
}
