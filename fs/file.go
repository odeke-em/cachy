package cachy

import (
	"io"
	"os"
	"time"
)

type File struct {
	uri     string
	isdir   bool
	path    string
	fsBased bool
	// ReadCloser Both response.Body and os.File implement io.ReadCloser
	// yet response.Body does not implement io.Seeker
	rc io.ReadCloser
	// Time-to-Live.
	ttl          int64
	dateModified time.Time
	closed       bool
}

func (f *File) Open() error {
	osf, err := os.Open(f.path)
	if err != nil {
		return err
	}
	f.rc = osf
	return nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := f.rc.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}

	return -1, ErrUnimplemented
}

func (f *File) Close() error {
	err := f.rc.Close()
	if err == nil && !f.closed {
		f.closed = true
	}
	return err
}

func (f *File) Read(b []byte) (int, error) {
	return f.rc.Read(b)
}

func (f *File) Ttl() int64 {
	return f.ttl
}
