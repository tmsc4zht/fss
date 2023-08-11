package fss

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"
)

type FS struct {
	fss map[string]fs.FS
}

func New() FS {
	return FS{fss: map[string]fs.FS{}}
}

func (fss FS) Add(name string, f fs.FS) {
	fss.fss[name] = f
}

func (fss FS) Open(name string) (fs.File, error) {

	if name == "" {
		return nil, fmt.Errorf("could not find path")
	}

	if name == "." {
		return rootFile{fss: &fss}, nil
	}

	v, p := separate(name)

	tfs, ok := fss.fss[v]
	if !ok {
		return nil, fmt.Errorf("could not find %s in fss", v)
	}

	return tfs.Open(p)

}

func (fss FS) ReadDir(name string) ([]fs.DirEntry, error) {
	ret := []fs.DirEntry{}

	if !fs.ValidPath(name) {
		return nil, os.ErrInvalid
	}

	if name == "." {
		for key, val := range fss.fss {
			ret = append(ret, dirEntry{name: key, fs: val})
		}

		sort.Slice(ret, func(i, j int) bool { return ret[i].Name() < ret[j].Name() })
		return ret, nil
	}

	v, p := separate(name)

	tfs, ok := fss.fss[v]
	if !ok {
		return ret, os.ErrNotExist
	}

	return fs.ReadDir(tfs, p)
}

type dirEntry struct {
	name string
	fs   fs.FS
}

func (d dirEntry) Info() (fs.FileInfo, error) {
	f, err := d.fs.Open(".")
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (d dirEntry) IsDir() bool {
	return true
}

func (d dirEntry) Name() string {
	return d.name
}

func (dirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

type rootFile struct {
	fss *FS
}

func (rootFile) Close() error {
	return nil
}

func (rootFile) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (rootFile) Stat() (fs.FileInfo, error) {
	return rootFileInfo{}, nil
}

func (r rootFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return r.fss.ReadDir(".")
}

type rootFileInfo struct {
}

func (rootFileInfo) IsDir() bool {
	return true
}

func (rootFileInfo) ModTime() time.Time {
	return time.Now()
}

func (rootFileInfo) Mode() fs.FileMode {
	return fs.ModeDir
}

func (rootFileInfo) Name() string {
	return "fss"
}

func (rootFileInfo) Size() int64 {
	return 0
}

func (rootFileInfo) Sys() any {
	return nil
}

func separate(name string) (string, string) {
	sp := strings.SplitN(name, "/", 2)
	if len(sp) == 1 {
		return sp[0], "."
	}
	return sp[0], sp[1]
}
