//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 16.11.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package fs

import (
	"net/http"
)

type Index404Fs struct {
	fs http.FileSystem
}

func (fs *Index404Fs) Open(name string) (http.File, error) {
	file, err := fs.fs.Open(name)
	if err != nil {
		return fs.fs.Open("/index.html")
	}
	return file, err
}
