package memfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/transientvariable/fs-go"

	gofs "io/fs"
	gohttp "net/http"
)

const (
	growthFactor = float32(1.618)
)

var (
	_ fs.File     = (*File)(nil)
	_ gohttp.File = (*File)(nil)
)

// File provides access to a single file or directory provided by MemFS.
//
// Implements the behavior defined by the fs.File and http.File interfaces.
type File struct {
	closed  bool
	dirIter fs.DirIterator
	fd      *fd
	flag    int
	mutex   sync.RWMutex
	rOff    int64
	wOff    int64
}

func newFile(fd *fd, flag int) (*File, error) {
	db := bytes.NewBuffer(fd.data)
	if flag&fs.O_TRUNC > 0 {
		db.Reset()
		fd.entry.SetSize(0)
	}
	return &File{fd: fd, flag: flag}, nil
}

func (f *File) Close() error {
	if f == nil {
		return gofs.ErrInvalid
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !f.closed {
		f.closed = true
		return nil
	}
	return fmt.Errorf("memfs_file: %w", &gofs.PathError{Op: "close", Err: gofs.ErrClosed})
}

func (f *File) Read(b []byte) (int, error) {
	fi, err := f.checkRead("read")
	if err != nil {
		return 0, err
	}

	if len(b) == 0 {
		return 0, nil
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.rOff >= fi.Size() {
		return 0, io.EOF
	}
	n := copy(b, f.fd.bytes()[f.rOff:])
	f.rOff += int64(n)
	return n, nil
}

func (f *File) ReadAt(b []byte, off int64) (int, error) {
	if _, err := f.checkRead("readAt"); err != nil {
		return 0, err
	}

	if len(b) == 0 {
		return 0, nil
	}

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	n := copy(b, f.fd.bytes()[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (f *File) ReadFrom(r io.Reader) (int64, error) {
	fi, err := f.checkWrite("readFrom")
	if err != nil {
		return 0, err
	}

	if r == nil {
		return 0, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "readFrom",
			Path: fi.Name(),
			Err:  errors.New("reader is nil"),
		})
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return n, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "readFrom",
			Path: fi.Name(),
			Err:  err,
		})
	}
	return n, nil
}

func (f *File) Readdir(count int) ([]gofs.FileInfo, error) {
	de, err := f.readDir(count)
	entries := make([]gofs.FileInfo, len(de))
	for i, e := range de {
		entries[i] = e
	}
	return entries, err
}

func (f *File) ReadDir(n int) ([]gofs.DirEntry, error) {
	de, err := f.readDir(n)
	entries := make([]gofs.DirEntry, len(de))
	for i, e := range de {
		entries[i] = e
	}
	return entries, err
}

func (f *File) Seek(off int64, whence int) (int64, error) {
	fi, err := f.checkRead("seek")
	if err != nil {
		return 0, err
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = off
	case io.SeekCurrent:
		abs = f.rOff + off
	case io.SeekEnd:
		abs = fi.Size() + off
	default:
		return 0, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "seek",
			Path: fi.Name(),
			Err:  errors.New("invalid whence"),
		})
	}

	if abs < 0 {
		return 0, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "seek",
			Path: fi.Name(),
			Err:  errors.New("negative position"),
		})
	}
	f.rOff = abs
	return abs, nil
}

func (f *File) Stat() (gofs.FileInfo, error) {
	if f == nil {
		return nil, gofs.ErrInvalid
	}

	if f.closed {
		return nil, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "stat",
			Path: f.fd.entry.Path(),
			Err:  gofs.ErrClosed,
		})
	}

	if f.fd.entry.Name() == "." {
		return f.fd.dir.entry, nil
	}
	return f.fd.entry, nil
}

func (f *File) Sync() error {
	return nil
}

func (f *File) Write(p []byte) (int, error) {
	if _, err := f.checkWrite("write"); err != nil {
		return 0, err
	}

	f.fd.mutex.Lock()
	defer f.fd.mutex.Unlock()

	if err := f.grow(len(p)); err != nil {
		return 0, err
	}

	n := copy(f.fd.data[f.wOff:], p)
	f.wOff += int64(n)

	if err := f.fd.entry.SetModTime(time.Now()); err != nil {
		return n, err
	}
	f.fd.entry.SetSize(uint64(f.wOff))
	return n, nil
}

// String returns a string representation of a File.
func (f *File) String() string {
	return ""
}

func (f *File) checkRegularFile(op string) (gofs.FileInfo, error) {
	fi, err := f.Stat()
	if err != nil {
		return fi, err
	}

	if fi.IsDir() {
		return fi, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   op,
			Path: fi.Name(),
			Err:  errors.New("is a directory"),
		})
	}
	return fi, nil
}

func (f *File) checkRead(op string) (gofs.FileInfo, error) {
	fi, err := f.checkRegularFile(op)
	if err != nil {
		return fi, err
	}

	if f.flag == fs.O_WRONLY {
		return fi, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   op,
			Path: fi.Name(),
			Err:  errors.New("file is write-only"),
		})
	}
	return fi, nil
}

func (f *File) checkWrite(op string) (gofs.FileInfo, error) {
	fi, err := f.checkRegularFile(op)
	if err != nil {
		return fi, err
	}

	if f.flag == fs.O_RDONLY {
		return fi, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   op,
			Path: fi.Name(),
			Err:  errors.New("file is read-only"),
		})
	}
	return fi, nil
}

func (f *File) grow(n int) error {
	currentCap := cap(f.fd.data)
	if len(f.fd.data)+n >= currentCap {
		c := int(growthFactor * float32(currentCap+n))
		if c > fs.MaxContentLen-c-n {
			return fs.ErrTooLarge
		}
		n := make([]byte, c)
		copy(n, f.fd.data)
		f.fd.data = n
	}
	return nil
}

func (f *File) readDir(n int) ([]*fs.Entry, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !fi.IsDir() {
		return nil, fmt.Errorf("memfs_file: %w", &gofs.PathError{
			Op:   "readDir",
			Path: fi.Name(),
			Err:  fs.ErrNotDir,
		})
	}

	if f.dirIter == nil {
		f.dirIter = newDirIterator(f.fd.dir)
	}
	return f.dirIter.NextN(n)
}
