package fs

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/transientvariable/anchor"

	json "github.com/json-iterator/go"
	gofs "io/fs"
	gopath "path"
)

var (
	_ gofs.DirEntry = (*Entry)(nil)
	_ gofs.FileInfo = (*Entry)(nil)
)

type PathValidator func(string) bool

// Entry is a container for file and directory metadata.
type Entry struct {
	attrs         *Attribute
	path          string
	pathValidator PathValidator
}

// NewEntry creates a new Entry.
func NewEntry(path string, options ...func(*Entry)) (*Entry, error) {
	entry := &Entry{path: path}
	for _, opt := range options {
		opt(entry)
	}

	if entry.attrs == nil {
		attrs, err := NewAttributes()
		if err != nil {
			return entry, err
		}
		entry.attrs = attrs
	}

	if err := validPath(path, entry.pathValidator); err != nil {
		return entry, err
	}
	return entry, nil
}

// Attributes returns the attributes for the Entry.
func (e *Entry) Attributes() *Attribute {
	return e.attrs
}

// Dir returns the path for the Entry with the last element truncated.
func (e *Entry) Dir() string {
	return gopath.Dir(e.path)
}

// Info ...
func (e *Entry) Info() (gofs.FileInfo, error) {
	return e, nil
}

// IsDir returns whether the Entry represents a directory.
func (e *Entry) IsDir() bool {
	return e.attrs.mode&gofs.ModeDir != 0
}

// Mode returns mode bits for the Entry.
func (e *Entry) Mode() gofs.FileMode {
	return e.attrs.mode
}

// ModTime returns the modification time for the Entry.
func (e *Entry) ModTime() time.Time {
	return e.attrs.Mtime()
}

// Name returns the Entry name.
func (e *Entry) Name() string {
	return gopath.Base(e.path)
}

// Path returns the full path for the Entry.
func (e *Entry) Path() string {
	return e.path
}

// Size returns the length in bytes if an Entry represents a regular file.
func (e *Entry) Size() int64 {
	return e.attrs.size
}

// SetModTime sets the modification time for the Entry.
func (e *Entry) SetModTime(t time.Time) error {
	t = t.UTC()
	if t.IsZero() || t.Equal(e.attrs.mtime) {
		return nil
	}

	if t.Before(e.attrs.mtime) {
		return fmt.Errorf("entry: %w", ErrMtimeMismatch)
	}
	e.attrs.mtime = t
	return nil
}

// SetPath sets the path for the Entry.
func (e *Entry) SetPath(p string) error {
	if err := validPath(p, e.pathValidator); err != nil {
		return err
	}
	e.path = p
	return nil
}

// SetSize sets the size for the Entry if it represents a regular file.
func (e *Entry) SetSize(s uint64) {
	if !e.IsDir() {
		e.attrs.size = int64(s)
	}
}

// Sys returns the underlying data source for the Entry (can return nil).
func (e *Entry) Sys() any {
	return nil
}

// Type returns the type bits for the Entry.
//
// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
func (e *Entry) Type() gofs.FileMode {
	return e.Mode().Type()
}

// Copy returns a copy of the Entry.
func (e *Entry) Copy() *Entry {
	var attrs *Attribute
	if e.attrs != nil {
		attrs = e.attrs.Copy()
	}
	return &Entry{
		attrs:         attrs,
		path:          e.path,
		pathValidator: e.pathValidator,
	}
}

// ToMap returns a map representation of the Entry properties.
func (e *Entry) ToMap() (map[string]any, error) {
	var m map[string]any
	if err := json.NewDecoder(strings.NewReader(e.String())).Decode(&m); err != nil {
		return m, err
	}
	return m, nil
}

// String returns a string representation of the Entry.
func (e *Entry) String() string {
	s := make(map[string]any)
	s["dir"] = e.Dir()
	s["is_dir"] = e.IsDir()
	s["name"] = e.Name()
	s["mode"] = e.Mode().String()
	s["mod_time"] = e.ModTime()
	s["path"] = e.Path()
	s["size"] = e.Size()
	s["type"] = e.Type()

	if e.attrs != nil {
		attrs, err := e.attrs.ToMap()
		if err != nil {
			s["attributes"] = err.Error()
		} else {
			s["attributes"] = attrs
		}
	}
	return string(anchor.ToJSONFormatted(s))
}

func validPath(p string, v func(string) bool) error {
	if v == nil {
		v = gofs.ValidPath
	}

	if !v(p) {
		return errors.New(fmt.Sprintf("entry: path is invalid: %s", p))
	}
	return nil
}

// WithAttributes sets the Attribute for an Entry.
func WithAttributes(attrs *Attribute) func(*Entry) {
	return func(e *Entry) {
		e.attrs = attrs
	}
}

// WithPathValidator sets the function used for validating paths for the Entry.
func WithPathValidator(v func(string) bool) func(*Entry) {
	return func(e *Entry) {
		e.pathValidator = v
	}
}
