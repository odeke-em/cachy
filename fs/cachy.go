package cachy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/odeke-em/cachy/urifspather"
)

const (
	defaultRoot = "/cachy"
)

var (
	ErrUnimplemented = errors.New("unimplemented")
)

type FileSystem struct {
	mu           sync.Mutex
	storageGuard atomic.Value
	root         string
}

type ReadSeekCloser interface {
	io.Reader
	io.Closer
	io.Seeker
}

func (fs *FileSystem) Open(uri string) (ReadSeekCloser, error) {
	cache, err := fs.loadCache()
	if err != nil {
		return nil, err
	}

	if retr, ok := cache[uri]; ok {
		log.Printf("cache hit for %q, %+v\n", uri, retr)
		retr.Open()
		return retr, nil
	}

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.Close && resp.Body != nil {
		defer resp.Body.Close()
	}

	fsTranslator, err := fser.New(uri)
	if err != nil {
		return nil, fmt.Errorf("uri %q, creating a uriTofs path resolver, got err %v", uri, err)
	}

	dirpath := fsTranslator.Dirname()

	// Commit it to memory
	err = os.MkdirAll(dirpath, 0755)
	if err != nil {
		return nil, fmt.Errorf("mkdirAll: err %v; path %q", err, dirpath)
	}

	fullOutpath := filepath.Clean(filepath.Join(fs.root, fsTranslator.Path()))
	osF, err := os.Create(fullOutpath)
	if err != nil {
		return nil, fmt.Errorf("err %v; creating outpath %q for uri %q", err, fullOutpath, uri)
	}

	n, err := io.Copy(osF, resp.Body)

	if err != nil && n < 1 { // Taking into account retryable io.UnexpectedEOF errors
		// Rollback
		_ = osF.Close()
		_ = os.Remove(fullOutpath)
		return nil, fmt.Errorf("err %v; copying body to %q for uri %q", err, fullOutpath, uri)
	}

	// Let it be committed to memory and closed
	_ = osF.Close()

	// TODO: @odeke-em read off the headers' "Date/Last-Modified"
	// and then  Chtime the outpath to make dates correspond.

	log.Println("done creating", fullOutpath)
	f := &File{
		uri:     uri,
		path:    fullOutpath,
		fsBased: true,
		ttl:     -1,
	}

	_ = f.Open()

	newCache := make(map[string]*File)
	for k, v := range cache {
		newCache[k] = v
	}

	newCache[f.uri] = f

	_ = fs.storeCache(newCache)
	fmt.Println("cache", newCache)

	return f, nil
}

func New(root string) (*FileSystem, error) {
	fs := &FileSystem{
		root: root,
	}
	fs.init()
	return fs, nil
}

func (fs *FileSystem) init() {
	fs.root = strings.TrimSpace(fs.root)
	if fs.root == "" {
		fs.root = defaultRoot
	}
	cache := make(map[string]*File)
	fs.storageGuard.Store(cache)
}

func explore(rootPath string) (chan *File, error) {
	fChan := make(chan *File)

	rootOSF, err := ioutil.ReadDir(rootPath)
	if err != nil {
		close(fChan)
		return fChan, err
	}

	chanOChans := make(chan chan *File)
	go func() {
		defer close(chanOChans)

		for _, childOSF := range rootOSF {
			fileName := childOSF.Name()
			fullChildPath := filepath.Join(rootPath, fileName)
			f := &File{
				dateModified: childOSF.ModTime(),
				path:         fullChildPath,
				rc:           nil,
				fsBased:      true,
				ttl:          -1,
			}

			fChan <- f

			if childOSF.IsDir() {
				cfChan, _ := explore(fullChildPath)
				chanOChans <- cfChan
			}
		}
	}()

	go func() {
		defer close(fChan)
		for chChan := range chanOChans {
			for chF := range chChan {
				fChan <- chF
			}
		}

	}()

	return fChan, nil
}

func (fs *FileSystem) Mount() (string, error) {
	fs.init()

	rootPath := fs.root
	err := os.MkdirAll(rootPath, 0755)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// TODO: Clean up on last error

	cache, err := fs.loadCache()
	if err != nil {
		return "", err
	}

	rootOSF, err := explore(rootPath)
	if err != nil {
		return "", err
	}

	for f := range rootOSF {
		cache[f.path] = f
	}

	if err := fs.storeCache(cache); err != nil {
		return "", fmt.Errorf("saving cache, err %v", err)
	}

	return rootPath, nil
}

func (fs *FileSystem) loadCache() (map[string]*File, error) {
	retr := fs.storageGuard.Load()
	if mapping, ok := retr.(map[string]*File); ok {
		return mapping, nil
	}
	return make(map[string]*File), nil
}

func (fs *FileSystem) storeCache(m map[string]*File) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.storageGuard.Store(m)
	return nil
}
