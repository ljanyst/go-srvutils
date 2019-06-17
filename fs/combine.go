//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 16.06.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package fs

import (
	"net/http"
	"os"
	"path"
	"strings"
)

type ComposedEntry struct {
	MountPoint string
	Fs         http.FileSystem
}

type ComposedDir struct {
	http.File
	extraEntries []string
}

type ComposedFileSystem struct {
	root        http.FileSystem
	mountPoints map[string]http.FileSystem
}

func (d ComposedDir) Readdir(count int) ([]os.FileInfo, error) {
	entries, err := d.File.Readdir(count)
	if err != nil {
		return entries, nil
	}

	for _, name := range d.extraEntries {
		entries = append(entries, VirtualFile{name, 0, true, nil})
	}
	return entries, nil
}

func (fs ComposedFileSystem) Open(name string) (http.File, error) {
	if name == "/" {
		file, err := fs.root.Open(name)

		if err != nil {
			return nil, err
		}

		var subDirs []string
		for name, _ := range fs.mountPoints {
			subDirs = append(subDirs, name)
		}
		return ComposedDir{file, subDirs}, nil
	}

	components := strings.Split(name, "/")

	if len(components) < 2 || !strings.HasPrefix(name, "/") {
		return fs.root.Open(name)
	}

	if subFs, ok := fs.mountPoints[components[1]]; ok {
		return subFs.Open(path.Join(components[2:]...))
	}

	return fs.root.Open(name)
}

func NewComposedFileSystem(root http.FileSystem, entries []ComposedEntry) http.FileSystem {
	mountPoints := make(map[string]http.FileSystem)
	for _, entry := range entries {
		mountPoints[entry.MountPoint] = entry.Fs
	}
	return ComposedFileSystem{root, mountPoints}
}
