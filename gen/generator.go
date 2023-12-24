//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 10.06.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package gen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type projectOpts struct {
	Name string
}

func isProgramAvailable(name string) bool {
	cmd := exec.Command(name, "-v")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

type Options struct {
	ProjectPath     string
	PostInstallHook func() error
	BuildProject    bool
	Assets          http.FileSystem
	PackageName     string
	BuildTags       string
	VariableName    string
	Filename        string
}

func GenerateNodeProject(opts Options) error {
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})

	log.Infof("Working in: %s", opts.ProjectPath)

	// Check the validity of the project
	jsonFile := filepath.Join(opts.ProjectPath, "package.json")
	log.Infof("Reading project description from: %s", jsonFile)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("Unable to read the configuration file: %v", err)
	}

	var pOpts projectOpts
	err = json.Unmarshal(data, &pOpts)
	if err != nil {
		return fmt.Errorf("Malformed config: %v", err)
	}

	if pOpts.Name == "" {
		return fmt.Errorf("The project has no name!")
	}

	log.Infof("The name of the project is: %s", pOpts.Name)

	// Check if we have npm available
	if !isProgramAvailable("npm") {
		return fmt.Errorf("npm does not seem to be installed!")
	}

	// Clean up after previous builds
	log.Info("Removing data of the previous build...")
	for _, dir := range []string{"dist", "node_modules"} {
		if err = os.RemoveAll(filepath.Join(opts.ProjectPath, dir)); os.IsExist(err) {
			return fmt.Errorf("Cannot remove %s: %s", dir, err)
		}
	}

	// Install node packages
	log.Info("Installing the JavaScript dependencies...")
	cmd := exec.Command("npm", "install")
	cmd.Dir = opts.ProjectPath
	if err = cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		return fmt.Errorf("Cannot install the JavaScript dependencies:\n%s", string(output))
	}

	// Run the post install hook
	if opts.PostInstallHook != nil {
		log.Info("Running the post install hook...")
		if err := opts.PostInstallHook(); err != nil {
			return fmt.Errorf("The post install hook failed: %s", err)
		}
	}

	// Build the UI assets
	if opts.BuildProject {
		log.Info("Building the UI assets...")
		cmd = exec.Command("npm", "run-script", "build")
		cmd.Dir = opts.ProjectPath
		if err = cmd.Run(); err != nil {
			output, _ := cmd.CombinedOutput()
			return fmt.Errorf("Cannot build UI assets:\n%s", string(output))
		}
	}

	// Generate the asset file
	log.Info("Generating the asset file...")
	err = vfsgen.Generate(opts.Assets, vfsgen.Options{
		PackageName:  opts.PackageName,
		BuildTags:    opts.BuildTags,
		VariableName: opts.VariableName,
		Filename:     opts.Filename,
	})

	if err != nil {
		return err
	}

	log.Info("All done.")
	return nil
}
