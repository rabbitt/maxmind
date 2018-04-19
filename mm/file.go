package mm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/yookoala/realpath"
)

const Separator = string(os.PathSeparator)

// ffjson: skip
type Pathname struct {
	path string
}

func NewPathname(path string) *Pathname {
	return &Pathname{path: path}
}

func (f *Pathname) Path() string { return f.path }

func (f *Pathname) String() string {
	if f != nil {
		return f.path
	} else {
		return ""
	}
}

func (f *Pathname) Set(value string) error {
	if value == "" {
		return errors.New("invalid path; path empty!")
	} else {
		*f = Pathname{path: value}
	}
	return nil
}

func (f *Pathname) RealPath() (*Pathname, error) {
	var err error
	var fpath string

	// normalize the database file path
	if fpath, err = realpath.Realpath(f.path); err != nil {
		if !f.Exists() {
			err = errors.New(fmt.Sprintf("%s path does not exist", f.path))
		}
		return f, err
	}

	return NewPathname(fpath), err
}

func (f *Pathname) Join(paths ...interface{}) *Pathname {
	var finalPath = f.path

	for _, obj := range paths {
		var pathStruct *Pathname
		var pathString string
		var ok bool
		_, _ = pathStruct, pathString

		switch obj.(type) {
		case Pathname:
			if pathStruct, ok = obj.(*Pathname); ok {
				finalPath = filepath.Join(finalPath, pathStruct.path)
			}
		case string:
			if pathString, ok = obj.(string); ok {
				finalPath = filepath.Join(finalPath, pathString)
			}
		default:
			continue
		}
	}
	return NewPathname(finalPath)
}

func (f *Pathname) Dirname() *Pathname {
	return NewPathname(filepath.Dir(f.path))
}

func (f *Pathname) Basename() *Pathname {
	return NewPathname(filepath.Base(f.path))
}

func (f *Pathname) Read() ([]byte, error) {
	return ioutil.ReadFile(f.path)
}

func (f *Pathname) Exists() bool {
	if f.path == "" {
		return false
	}
	_, err := os.Stat(f.path)
	if err != nil && os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
