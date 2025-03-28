package memfs

import (
	"github.com/transientvariable/fs-go"
	"github.com/transientvariable/log-go"
	"github.com/transientvariable/support-go"

	gofs "io/fs"
)

type fsEntry struct {
	entry *fs.Entry
	data  any
}

// Stat returns the FileInfo.
func (f *fsEntry) Stat() (gofs.FileInfo, error) {
	if f.entry != nil {
		return f.entry, nil
	}
	return nil, gofs.ErrInvalid
}

// Path returns the path to the File or MemFS.
func (f *fsEntry) Path() string {
	if f.entry != nil {
		return f.entry.Path()
	}
	return ""
}

// Value satisfies the Trie.Entry interface Value() method. The value will be the path to the File or MemFS type.
func (f *fsEntry) Value() string {
	return f.Path()
}

// Data satisfies the Trie.Entry interface Data() method. The type returned with either be a File or MemFS type.
func (f *fsEntry) Data() any {
	return f.data
}

// String returns a string representation of the fsEntry.
func (f *fsEntry) String() string {
	if f.entry != nil {
		s := make(map[string]any)
		s["path"] = f.entry.Path()

		a, err := f.entry.Attributes().ToMap()
		if err != nil {
			log.Error("[memfs:entry]", log.Err(err))
		}

		s["entry"] = map[string]any{
			"attributes": a,
		}
		return string(support.ToJSONFormatted(s))
	}
	return ""
}
