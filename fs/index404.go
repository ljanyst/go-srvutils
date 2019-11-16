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
	Fs http.FileSystem
}

func (fs *Index404Fs) Open(name string) (http.File, error) {
	file, err := fs.Fs.Open(name)
	if err != nil {
		return fs.Fs.Open("/index.html")
	}
	return file, err
}
