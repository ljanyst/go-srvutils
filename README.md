go-srvutils
===========

This project is a container for a bunch of utilities for serving web apps with
go. They address the pain points that are not necessarily caused by go itself,
but also by the sorry state of the JavaScript environment.

nodeinstaller
-------------

Node installer is a script that fetches the `nodejs` source code, compiles it,
installs in a directory of your choosing, and provides an activation script
similar to the one of Python's virtualenv.

    go get github.com/ljanyst/go-srvutils/...
    go run github.com/ljanyst/go-srvutils/nodeinstaller -version=10.16.0 \
        -prefix=/some/absolute/installation/prefix

You can then activate this environment whenever you want using the
aforementioned activation script:

    . /some/absolute/installation/prefix/bin/activate

fs
--

The `fs` package provides functionality that is not really necessary if you just
want to build a go app that serves files from a filesystem, but comes handy when
you want to create a self-contained single binary app. It may be used with
generators such as `vfsgen` to combine multiple directories and filter the files
that you want to embed in the binary.

```go
// +build dev

package webpass2

import (
	"net/http"

	"github.com/ljanyst/go-srvutils/fs"
)

var Assets http.FileSystem = fs.NewFilteredFileSystem(
	fs.NewComposedFileSystem(
		http.Dir("../../ui/public"),
		[]fs.ComposedEntry{
			{"node_modules", http.Dir("../../ui/node_modules")},
		},
	),
	[]string{
		"/index.html",
		"/node_modules/onsenui/css/onsenui.css",
		"/node_modules/onsenui/css/onsen-css-components.css",
		"/node_modules/vue/dist/vue.min.js",
		"/node_modules/vue-onsenui/dist/vue-onsenui.min.js",
		"/node_modules/onsenui/js/onsenui.min.js",
		"/node_modules/openpgp/dist/openpgp.min.js",
		"/main.js",
		"/favicon.png",
		"/node_modules/onsenui/css/ionicons/css/ionicons.min.css",
		"/node_modules/onsenui/css/material-design-iconic-font/css/material-design-iconic-font.min.css",
		"/node_modules/onsenui/css/font_awesome/css/font-awesome.min.css",
		"/node_modules/onsenui/css/font_awesome/css/v4-shims.min.css",
		"/node_modules/openpgp/dist/openpgp.worker.min.js",
	},
)
```

gen
---

Oftentimes you need to install and build an `npm` project before you can embed
the resulting files into a binary. The `gen` package helps with that.

```go
// +build ignore

package main

import (
	"github.com/ljanyst/go-srvutils/gen"
	"github.com/ljanyst/webpass2/pkg/webpass2"
	"log"
)

func main() {
	err := gen.GenerateNodeProject(gen.Options{
		ProjectPath:  "../../ui",
		BuildProject: false,
		Assets:       webpass2.Assets,
		PackageName:  "webpass2",
		BuildTags:    "!dev",
		VariableName: "Assets",
		Filename:     "assets_prod.go",
	})

	if err != nil {
		log.Fatal(err)
	}
}
```

You can then use it with `go generate`:

```go
//go:generate go run --tags=dev assets_generate.go
```