//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 20.06.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// The activator code is taken from nodeenv by Eugene Kalinin
// https://github.com/ekalinin/nodeenv/blob/master/nodeenv.py
// Distributed under the MIT License
const activator string = `
# This is taken from https://github.com/ekalinin/nodeenv/blob/master/nodeenv.py
# And massaged slightly
# MIT Licensed

# This file must be used with "source bin/activate" *from bash*
# you cannot run it directly

deactivate_node () {
    # reset old environment variables
    if [ -n "$_OLD_NODE_VIRTUAL_PATH" ] ; then
        PATH="$_OLD_NODE_VIRTUAL_PATH"
        export PATH
        unset _OLD_NODE_VIRTUAL_PATH

        NODE_PATH="$_OLD_NODE_PATH"
        export NODE_PATH
        unset _OLD_NODE_PATH

        NPM_CONFIG_PREFIX="$_OLD_NPM_CONFIG_PREFIX"
        npm_config_prefix="$_OLD_npm_config_prefix"
        export NPM_CONFIG_PREFIX
        export npm_config_prefix
        unset _OLD_NPM_CONFIG_PREFIX
        unset _OLD_npm_config_prefix
    fi

    # This should detect bash and zsh, which have a hash command that must
    # be called to get it to forget past commands.  Without forgetting
    # past commands the $PATH changes we made may not be respected
    if [ -n "$BASH" -o -n "$ZSH_VERSION" ] ; then
        hash -r
    fi

    if [ -n "$_OLD_NODE_VIRTUAL_PS1" ] ; then
        PS1="$_OLD_NODE_VIRTUAL_PS1"
        export PS1
        unset _OLD_NODE_VIRTUAL_PS1
    fi

    unset NODE_VIRTUAL_ENV
    if [ ! "$1" = "nondestructive" ] ; then
    # Self destruct!
        unset -f deactivate_node
    fi
}

# unset irrelevant variables
deactivate_node nondestructive

# find the directory of this script
# http://stackoverflow.com/a/246128
if [ "${BASH_SOURCE}" ] ; then
    SOURCE="${BASH_SOURCE[0]}"

    while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
    DIR="$( command cd -P "$( dirname "$SOURCE" )" > /dev/null && pwd )"

    NODE_VIRTUAL_ENV="$(dirname "$DIR")"
else
    # dash not movable. fix use case:
    #   dash -c " . node-env/bin/activate && node -v"
    NODE_VIRTUAL_ENV="__NODE_VIRTUAL_ENV__"
fi

# NODE_VIRTUAL_ENV is the parent of the directory where this script is
export NODE_VIRTUAL_ENV

_OLD_NODE_VIRTUAL_PATH="$PATH"
PATH="$NODE_VIRTUAL_ENV/lib/node_modules/.bin:$NODE_VIRTUAL_ENV/bin:$PATH"
export PATH

_OLD_NODE_PATH="$NODE_PATH"
NODE_PATH="$NODE_VIRTUAL_ENV/lib/modules"
export NODE_PATH

_OLD_NPM_CONFIG_PREFIX="$NPM_CONFIG_PREFIX"
_OLD_npm_config_prefix="$npm_config_prefix"
NPM_CONFIG_PREFIX="$NODE_VIRTUAL_ENV"
npm_config_prefix="$NODE_VIRTUAL_ENV"
export NPM_CONFIG_PREFIX
export npm_config_prefix

if [ -z "$NODE_VIRTUAL_ENV_DISABLE_PROMPT" ] ; then
    _OLD_NODE_VIRTUAL_PS1="$PS1"
    if [ "x__NODE_VIRTUAL_PROMPT__" != x ] ; then
        PS1="__NODE_VIRTUAL_PROMPT__ $PS1"
    else
    if [ "` + "`" + `basename \"$NODE_VIRTUAL_ENV\"` + "`" + `" = "__" ] ; then
        # special case for Aspen magic directories
        # see http://www.zetadev.com/software/aspen/
        PS1="[` + "`" + `basename \` + "`" + `dirname \"$NODE_VIRTUAL_ENV\"\` + "``" + `] $PS1"
    else
        PS1="(` + "`" + `basename \"$NODE_VIRTUAL_ENV\"` + "`" + `) $PS1"
    fi
    fi
    export PS1
fi

# This should detect bash and zsh, which have a hash command that must
# be called to get it to forget past commands.  Without forgetting
# past commands the $PATH changes we made may not be respected
if [ -n "$BASH" -o -n "$ZSH_VERSION" ] ; then
    hash -r
fi
`

func main() {
	// Commandline
	version := flag.String("version", "10.16.0", "version to install")
	cleanup := flag.Bool("cleanup", true, "clean up the temporary data")
	prefix := flag.String("prefix", "", "installation prefix")
	buildThreads := flag.Uint("threads", 12, "Number of build threads")
	flag.Parse()

	// Logging
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})

	// Check the params
	if *prefix == "" || (*prefix)[0] != '/' {
		log.Fatal("You need to specify an absolute installation prefix. Use the -prefix option.")
	}

	// Tempdir
	dir, err := ioutil.TempDir("", "node-inst")
	if err != nil {
		log.Fatalf("Unable to create the temdir: %s", err)
	}

	if *cleanup {
		defer os.RemoveAll(dir)
	}

	// Download
	nodeArchive := filepath.Join(dir, "node.tar.gz")
	log.Infof("Downloading node %s to %s", *version, nodeArchive)

	url := fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s.tar.gz", *version, *version)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Can't get the url %s: %s", url, err)
	}
	defer resp.Body.Close()

	out, err := os.Create(nodeArchive)
	if err != nil {
		log.Fatalf("Can't create the output file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Can't write the file body: %s", err)
	}

	// Unpack
	log.Infof("Unpacking...")
	cmd := exec.Command("tar", "zxf", nodeArchive)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		log.Fatalf("Cannot unpack node:\n%s\n%s", err, string(output))
	}

	// Configure
	log.Infof("Configuring...")
	nodeSource := filepath.Join(dir, fmt.Sprintf("node-v%s", *version))
	cmd = exec.Command("./configure", fmt.Sprintf("--prefix=%s", *prefix))
	cmd.Dir = nodeSource
	if err = cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		log.Fatalf("Cannot configure node:\n%s\n%s", err, string(output))
	}

	// Build
	log.Infof("Building...")
	cmd = exec.Command("make", fmt.Sprintf("-j%d", *buildThreads))
	cmd.Dir = nodeSource
	if err = cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		log.Fatalf("Cannot build node:\n%s\n%s", err, string(output))
	}

	// Install
	log.Infof("Installing...")
	cmd = exec.Command("make", "install")
	cmd.Dir = nodeSource
	if err = cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		log.Fatalf("Cannot install node:\n%s\n%s", err, string(output))
	}

	// Install activator
	finalActivator := strings.Replace(activator, "__NODE_VIRTUAL_PROMPT__",
		fmt.Sprintf("(node-%s)", *version), -1)
	finalActivator = strings.Replace(finalActivator, "__NODE_VIRTUAL_ENV__",
		*prefix, -1)

	log.Infof("Installing the activator script...")
	activatorPath := filepath.Join(*prefix, "bin", "activate")
	if err := ioutil.WriteFile(activatorPath, []byte(finalActivator), 0644); err != nil {
		log.Fatalf("Cannot install the activator: %s", err)
	}

	log.Infof("You can acrivate node by including the activator script in your shell environment:")
	log.Infof(". %s/bin/activate", *prefix)
}
