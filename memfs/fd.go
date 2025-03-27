package memfs

import (
	"errors"
	"sync"

	"github.com/transientvariable/fs"
	"github.com/transientvariable/log"

	gofs "io/fs"
)

// fd (file descriptor) represents File content and its associated metadata.
type fd struct {
	data  []byte
	dir   *MemFS
	entry *fs.Entry
	mutex sync.RWMutex
}

func newfd(dir *MemFS, name string, flag int, mode gofs.FileMode) (*fd, error) {
	e, err := entry(dir, name)
	if err != nil {
		if errors.Is(err, gofs.ErrNotExist) && flag&fs.O_CREATE != 0 {
			log.Trace("[memfs:fd] creating new file descriptor",
				log.String("directory", dir.entry.Name()),
				log.String("name", name),
			)

			attrs, err := fs.NewAttributes(fs.WithMode(uint32(mode)))
			if err != nil {
				return nil, err
			}

			e, err := fs.NewEntry(name, fs.WithAttributes(attrs))
			if err != nil {
				return nil, err
			}

			fd := &fd{entry: e, dir: dir}
			if err := dir.entries.AddEntry(&fsEntry{entry: e, data: fd}); err != nil {
				return nil, err
			}
			return fd, nil
		}
		return nil, err
	}

	switch e.Data().(type) {
	case *fd:
		return e.Data().(*fd), nil
	case *MemFS:
		e, err := entry(e.Data().(*MemFS), ".")
		if err != nil {
			return nil, err
		}
		return e.Data().(*fd), nil
	default:
		return nil, fs.ErrNotFile
	}
}

func (d *fd) bytes() []byte {
	d.mutex.RLock()
	defer d.mutex.RLock()

	if d.entry.Size() > 0 {
		return d.data[:d.entry.Size()]
	}
	return d.data
}
