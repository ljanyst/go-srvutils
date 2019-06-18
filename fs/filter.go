//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 17.06.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package fs

import (
	"net/http"
	"os"
	"strings"
)

type FsNode struct {
	Children map[string]FsNode
}

type FilteredDir struct {
	http.File
	node FsNode
}

type FilteredFileSystem struct {
	fs        http.FileSystem
	whitelist FsNode
}

func (d FilteredDir) Readdir(count int) ([]os.FileInfo, error) {
	lst, err := d.File.Readdir(count)
	var filtered []os.FileInfo
	for _, el := range lst {
		if _, ok := d.node.Children[el.Name()]; ok {
			filtered = append(filtered, el)
		}
	}
	return filtered, err
}

func (fs FilteredFileSystem) Open(path string) (http.File, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, os.ErrNotExist
	}

	components := strings.Split(path, "/")

	if path == "/" {
		components = components[1:]
	}

	if !hasComponents(fs.whitelist, components[1:]) {

		return nil, os.ErrNotExist
	}

	file, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		node, _ := getNode(fs.whitelist, components[1:])
		return FilteredDir{file, node}, nil
	}
	return file, nil
}

func addPath(node FsNode, components []string) {
	if len(components) == 0 {
		return
	}

	if _, ok := node.Children[components[0]]; !ok {
		node.Children[components[0]] = FsNode{make(map[string]FsNode)}
	}

	addPath(node.Children[components[0]], components[1:])
}

func hasComponents(node FsNode, components []string) bool {
	if len(components) == 0 {
		return true
	}

	child, ok := node.Children[components[0]]
	if !ok {
		return false
	}
	return hasComponents(child, components[1:])
}

func getNode(node FsNode, components []string) (FsNode, error) {
	if len(components) == 0 {
		return node, nil
	}

	child, ok := node.Children[components[0]]
	if !ok {
		return FsNode{}, nil
	}
	return getNode(child, components[1:])
}

// Build new filtered filesystem. The whitelisted paths must start with a leading slash.
func NewFilteredFileSystem(fs http.FileSystem, whitelist []string) http.FileSystem {
	filteredFs := FilteredFileSystem{fs, FsNode{}}
	filteredFs.whitelist.Children = make(map[string]FsNode)
	for _, path := range whitelist {
		if !strings.HasPrefix(path, "/") {
			continue
		}
		components := strings.Split(path, "/")
		addPath(filteredFs.whitelist, components[1:])
	}
	return filteredFs
}
