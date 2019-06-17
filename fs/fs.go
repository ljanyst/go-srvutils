//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 11.06.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package fs

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

type VirtualFile struct {
	name  string
	size  int64
	isDir bool
	*bytes.Reader
}

func (f VirtualFile) Readdir(count int) ([]os.FileInfo, error) {
	if f.isDir {
		return []os.FileInfo{}, nil
	}
	return nil, fmt.Errorf("cannot Readdir from a file")
}

func (f VirtualFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f VirtualFile) IsDir() bool {
	return f.isDir
}

func (f VirtualFile) ModTime() time.Time {
	return time.Now()
}

func (f VirtualFile) Mode() os.FileMode {
	if f.isDir {
		return 0555
	}
	return 0444
}

func (f VirtualFile) Size() int64 {
	return f.size
}

func (f VirtualFile) Close() error {
	return nil
}

func (f VirtualFile) Name() string {
	return f.name
}

func (f VirtualFile) Sys() interface{} {
	return nil
}
