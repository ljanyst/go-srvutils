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

The package also comes with the `VirtualFile` class that implements the
`http.File` interface. It lets you easily serve memory blobs.

```go
fs.VirtualFile{
        filepath.Base(path),
        int64(len(data)),
        false,
        bytes.NewReader(data),
}
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

auth
----

The go http package seems to lack a simple HTTP BasicAuth handler that would
authenticate users using a user-password map read from a htpasswd file.


```go
passwords, err := htpasswd.ParseHtpasswdFile(authFile)
if err != nil {
	log.Fatalf(`Authentication enabled but cannot open htpassword file "%s": %s`,	authFile, err)
}

http.Handle("/", auth.NewBasicAuthHandler("realm", passwords, s3Fs))
```

websocket
---------

Websocket is a library for receiving, processing in a synchronized manner,
responding, unicasting, and broadcasting message over webosckets. It unmarshals
JSON strings into predefined objects based on action types.

```go
import (
	ws "github.com/ljanyst/go-srvutils/websocket"
)

type BrowserGetRequest struct {
	ws.RequestHeader
	Path string
}

type SocketHandler struct {
	compiler *Compiler
}

func (h *SocketHandler) ProcessRequest(req ws.Request) []ws.Response {
	switch req.Action() {
	case "BROWSER_GET":
		r := req.(*BrowserGetRequest)
		listing, err := h.compiler.GetContent(r.Path, HTML_RENDERER)
		if err != nil {
			return []ws.Response{
				ws.Response{ws.STATUS, ws.ERROR, err.Error(), r.Id()},
			}
		}
		return []ws.Response{
			ws.Response{ws.STATUS, ws.OK, listing, r.Id()},
		}
	}
	return []ws.Response{}
}

func (h *SocketHandler) NewClient() []ws.Response {
	return []ws.Response{}
}
```

```go
wsHandler := SocketHandler{compiler}
requestMap := map[string]reflect.Type{}
requestMap["BROWSER_GET"] = reflect.TypeOf(&BrowserGetRequest{})

ws, err := ws.NewWebSocketHandler(&wsHandler, requestMap)
if err != nil {
	log.Fatalf("Can't create a web socket handler: %s", err)
}

http.Handle("/ws", ws)
```

shortuuidgen
------------

`shortuuidgen` generates concise UUIDs using the [`shortuuid`][suuid] library.

[suuid]: https://github.com/lithammer/shortuuid
