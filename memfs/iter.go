package memfs

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/transientvariable/collection"
	"github.com/transientvariable/fs"
)

type dirIterator struct {
	iter collection.Iterator[string]
	mfs  *MemFS
}

func newDirIterator(mfs *MemFS) fs.DirIterator {
	return &dirIterator{
		iter: mfs.entries.Iterate(),
		mfs:  mfs,
	}
}

// HasNext returns whether the directory has remaining entries.
func (i *dirIterator) HasNext() bool {
	return i.iter.HasNext()
}

// Next returns the next directory fs.Entry. Dot entries "." are skipped.
//
// The error io.EOF is returned if there are no remaining entries left to iterate.
func (i *dirIterator) Next() (*fs.Entry, error) {
	if !i.HasNext() {
		return nil, io.EOF
	}

	v, err := i.iter.Next()
	if err != nil {
		if errors.Is(err, collection.ErrNotFound) {
			return nil, io.EOF
		}
		return nil, err
	}

	if v == "." {
		return i.Next()
	}

	e, err := i.mfs.entries.Entry(v)
	if err != nil {
		return nil, err
	}

	switch e.Data().(type) {
	case *MemFS:
		return e.Data().(*MemFS).entry, nil
	case *fd:
		return e.Data().(*fd).entry, nil
	default:
		return nil, fmt.Errorf("dir_iterator: %s: %w", reflect.ValueOf(e.Data()).Type(), fs.ErrInvalidEntryType)
	}
}

// NextN returns a slice containing the next n directory entries. Dot entries "." are skipped.
//
// The error io.EOF is returned if there are no remaining entries left to iterate.
func (i *dirIterator) NextN(n int) ([]*fs.Entry, error) {
	var entries []*fs.Entry
	if n > 0 {
		for j := 0; j < n; j++ {
			e, err := i.Next()
			if err != nil {
				return entries, err
			}
			entries = append(entries, e)
		}
		return entries, nil
	}

	for i.HasNext() {
		e, err := i.Next()
		if err != nil {
			return entries, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
