package fser

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
)

var (
	ErrNonEmptyPathRequired = errors.New("non empty path required")
	ErrInvalidScheme        = errors.New("invalid scheme")
)

type URIFsPather struct {
	dirname  string
	basename string
	rawURI   string
}

func md5Sum(p string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(p)))
}

func New(uri string) (*URIFsPather, error) {
	if uri == "" {
		return nil, ErrNonEmptyPathRequired
	}
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("empty scheme for %q", uri)
	}
	dirp, basep := filepath.Split(parsedURL.Path)

	hashedDirpath := md5Sum(dirp)
	dpst := &URIFsPather{
		rawURI:   uri,
		basename: basep,
		dirname:  filepath.Join(parsedURL.Scheme, hashedDirpath),
	}

	return dpst, nil
}

func (d URIFsPather) Dirname() string {
	return d.dirname
}

func (d URIFsPather) Basename() string {
	return d.basename
}

func (d URIFsPather) Path() string {
	return filepath.Join(d.dirname, d.basename)
}

func (d URIFsPather) String() string {
	return d.Path()
}

func (d URIFsPather) RawURI() string {
	return d.rawURI
}
