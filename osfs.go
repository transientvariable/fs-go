package fs

import (
	"os"
	"path/filepath"
	"runtime"

	gofs "io/fs"
)

var (
	_ FS = (*OSFS)(nil)
)

// OSFS os/platform file system provider that implements FS.
type OSFS struct{}

// New creates a new OSFS.
func New() (*OSFS, error) {
	return &OSFS{}, nil
}

func (o *OSFS) Close() error {
	return nil
}

func (o *OSFS) Open(name string) (gofs.File, error) {
	return os.Open(name)
}

func (o *OSFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (o *OSFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *OSFS) ReadDir(name string) ([]gofs.DirEntry, error) {
	return os.ReadDir(name)
}

func (o *OSFS) Stat(name string) (gofs.FileInfo, error) {
	return os.Stat(name)
}

func (o *OSFS) Sub(dir string) (gofs.FS, error) {
	return gofs.Sub(o, dir)
}

func (o *OSFS) Create(name string) (File, error) {
	return os.Create(name)
}

func (o *OSFS) Mkdir(name string, perm gofs.FileMode) error {
	return os.Mkdir(name, perm)
}

func (o *OSFS) MkdirAll(path string, perm gofs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *OSFS) OpenFile(name string, flag int, perm gofs.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

func (o *OSFS) PathSeparator() string {
	// TODO: do we need build arg variant for winduz (e.g. // +build windows ...)?
	return string(os.PathSeparator)
}

func (o *OSFS) Provider() string {
	return runtime.GOOS
}

func (o *OSFS) Remove(name string) error {
	return os.Remove(name)
}

func (o *OSFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (o *OSFS) Rename(oldpath string, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (o *OSFS) Root() (string, error) {
	return o.PathSeparator(), nil
}

func (o *OSFS) WriteFile(name string, data []byte, perm gofs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
