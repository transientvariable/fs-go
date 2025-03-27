package fs

// Enumeration of errors that may be returned by file system operations.
const (
	ErrCtimeMismatch    = fsError("modification time occurs before creation time")
	ErrIsDir            = fsError("is a directory")
	ErrInvalidEntryType = fsError("entry type is invalid")
	ErrMtimeMismatch    = fsError("modification time is invalid")
	ErrNotDir           = fsError("not a directory")
	ErrNotFile          = fsError("not a file")
	ErrTooLarge         = fsError("too large")
)

// fsError defines the type for errors that may be returned by file system operations.
type fsError string

// Error returns the cause of the file system error.
func (e fsError) Error() string {
	return string(e)
}
