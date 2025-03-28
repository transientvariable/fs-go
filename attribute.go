package fs

import (
	"fmt"
	"strings"
	"time"

	"github.com/transientvariable/support-go"

	json "github.com/json-iterator/go"
	gofs "io/fs"
)

// Attribute ...
type Attribute struct {
	ctime    time.Time
	gid      int32
	group    string
	inode    int64
	mimeType string
	mode     gofs.FileMode
	mtime    time.Time
	owner    string
	size     int64
	uid      int32
}

// NewAttributes ..
func NewAttributes(attributes ...func(*Attribute)) (*Attribute, error) {
	attrs := &Attribute{}
	for _, attr := range attributes {
		attr(attrs)
	}

	if attrs.ctime.IsZero() {
		attrs.ctime = time.Now().UTC()
	}

	if mtime := attrs.mtime; !mtime.IsZero() && mtime.Before(attrs.ctime) {
		return nil, fmt.Errorf("attribute: %w", ErrCtimeMismatch)
	}
	return attrs, nil
}

// Ctime ...
func (a *Attribute) Ctime() time.Time {
	return a.ctime
}

// GID ...
func (a *Attribute) GID() int32 {
	return a.gid
}

// Group ...
func (a *Attribute) Group() string {
	return a.group
}

// Inode ...
func (a *Attribute) Inode() int64 {
	return a.inode
}

// MimeType ...
func (a *Attribute) MimeType() string {
	return a.mimeType
}

// Mode ...
func (a *Attribute) Mode() gofs.FileMode {
	return a.mode
}

// Mtime ...
func (a *Attribute) Mtime() time.Time {
	return a.mtime
}

// Owner ...
func (a *Attribute) Owner() string {
	return a.owner
}

// Size ...
func (a *Attribute) Size() int64 {
	return a.size
}

// UID ...
func (a *Attribute) UID() int32 {
	return a.uid
}

// Copy returns a copy of the Attribute.
func (a *Attribute) Copy() *Attribute {
	return &Attribute{
		ctime:    a.Ctime(),
		gid:      a.GID(),
		group:    a.Group(),
		inode:    a.Inode(),
		mimeType: a.MimeType(),
		mode:     a.Mode(),
		mtime:    a.Mtime(),
		owner:    a.Owner(),
		size:     a.Size(),
		uid:      a.UID(),
	}
}

// ToMap returns a map representation of the Attribute properties.
func (a *Attribute) ToMap() (map[string]any, error) {
	var m map[string]any
	if err := json.NewDecoder(strings.NewReader(a.String())).Decode(&m); err != nil {
		return m, err
	}
	return m, nil
}

// String returns a string representation of the Attribute properties.
func (a *Attribute) String() string {
	s := make(map[string]any)
	s["ctime"] = a.Ctime()
	s["gid"] = a.GID()
	s["group"] = a.Group()
	s["inode"] = a.Inode()
	s["mime_type"] = a.MimeType()
	s["mode"] = a.Mode()
	s["mtime"] = a.Mtime()
	s["owner"] = a.Owner()
	s["size"] = a.Size()
	s["uid"] = a.UID()
	return string(support.ToJSONFormatted(s))
}

// WithCtime ...
func WithCtime(ctime time.Time) func(*Attribute) {
	return func(a *Attribute) {
		a.ctime = ctime.UTC()
	}
}

// WithGID ...
func WithGID(gid uint32) func(*Attribute) {
	return func(attrs *Attribute) {
		attrs.gid = int32(gid)
	}
}

// WithGroup ...
func WithGroup(group string) func(*Attribute) {
	return func(a *Attribute) {
		a.group = group
	}
}

// WithInode ...
func WithInode(inode uint64) func(*Attribute) {
	return func(a *Attribute) {
		a.inode = int64(inode)
	}
}

// WithMimeType ...
func WithMimeType(mimeType string) func(*Attribute) {
	return func(a *Attribute) {
		a.mimeType = mimeType
	}
}

// WithMode ...
func WithMode(mode uint32) func(*Attribute) {
	return func(a *Attribute) {
		a.mode = gofs.FileMode(mode)
	}
}

// WithMtime ...
func WithMtime(mtime time.Time) func(*Attribute) {
	return func(a *Attribute) {
		a.mtime = mtime.UTC()
	}
}

// WithOwner ...
func WithOwner(owner string) func(*Attribute) {
	return func(a *Attribute) {
		a.owner = owner
	}
}

// WithSize ...
func WithSize(size uint64) func(*Attribute) {
	return func(a *Attribute) {
		a.size = int64(size)
	}
}

// WithUID ...
func WithUID(uid uint32) func(*Attribute) {
	return func(a *Attribute) {
		a.uid = int32(uid)
	}
}
