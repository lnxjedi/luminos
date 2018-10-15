// Copyright (c) 2012-2014 José Carlos Nieto, https://menteslibres.net/xiam
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"

	"github.com/lnxjedi/cli"
	"github.com/lnxjedi/to"
	"github.com/lnxjedi/yaml"
)

// Default values
const (
	envSettingsFile   = "./settings.yaml"
	envServerDomain   = "unix"
	envServerProtocol = "tcp"
)

// Global software settings.
var settings *yaml.Yaml

// Command line settings.
var flagSettings = flag.String("c", envSettingsFile, "Path to the settings.yaml file")
var flagIndex = flag.Bool("i", false, "Generate search index on start")

// runCommand is the structure that provides instructions for the "luminos
// run" subcommand.
type runCommand struct {
}

// Execute runs a luminos server using a settings file.
func (c *runCommand) Execute() (err error) {
	var stat os.FileInfo

	// If no settings file was specified, use the default.
	if *flagSettings == "" {
		*flagSettings = envSettingsFile
	}

	// Attempt to stat the settings file.
	stat, err = os.Stat(*flagSettings)

	// It must not return an error.
	if err != nil {
		return fmt.Errorf("error while opening %s: %q", *flagSettings, err)
	}

	// We must have a value in stat.
	if stat == nil {
		return fmt.Errorf("could not load settings file: %s", *flagSettings)
	}

	// And the file must not be a directory.
	if stat.IsDir() {
		return fmt.Errorf("could not open %s: it's a directory", *flagSettings)
	}

	// Now that we're positively sure that we have a valid file, let's try to
	// read settings from it.
	if settings, err = loadSettings(*flagSettings); err != nil {
		return fmt.Errorf("error while reading settings file %s: %q", *flagSettings, err)
	}

	if *flagIndex {
		for name, host := range hosts {
			content, err := host.GetContentPath()
			if err == nil {
				fmt.Printf("Host '%s' has content path '%s'\n", name, content)
				err := index(host.DocumentRoot, content)
				if err != nil {
					return fmt.Errorf("error indexing '%s': %v", name, err)
				}
			} else {
				return fmt.Errorf("error locating content path for '%s': %v", name, err)
			}
		}
	}

	// Starting settings watcher.
	if watch, err = settingsWatcher(); err == nil {
		watch.Add(*flagSettings)
	}

	// Reading setttings.
	serverType := to.String(settings.Get("server", "type"))

	domain := envServerDomain
	address := to.String(settings.Get("server", "socket"))

	if address == "" {
		domain = envServerProtocol
		address = fmt.Sprintf("%s:%d", to.String(settings.Get("server", "bind")), to.Int64(settings.Get("server", "port")))
	}

	// Creating a network listener.
	var listener net.Listener

	if listener, err = net.Listen(domain, address); err != nil {
		return fmt.Errorf("Could not create network listener: %q", err)
	}

	// Listener must be closed when the function exits.
	defer listener.Close()

	// Attempt to start a server.
	switch serverType {
	case "fastcgi":
		if err == nil {
			log.Printf("Starting FastCGI server. Listening at %s.\n", address)
			fcgi.Serve(listener, &server{})
		} else {
			return fmt.Errorf("Failed to start FastCGI server: %q", err)
		}
	case "standalone":
		if err == nil {
			log.Printf("Starting HTTP server. Listening at %s.\n", address)
			http.Serve(listener, &server{})
		} else {
			return fmt.Errorf("Failed to start HTTP server: %q", err)
		}
	default:
		return fmt.Errorf("Unknown server type: %s", serverType)
	}

	return nil
}

func init() {
	// Describing the "run" subcommand.

	cli.Register("run", cli.Entry{
		Name:        "run",
		Description: "Runs a luminos server.",
		Arguments:   []string{"c", "i"},
		Command:     &runCommand{},
	})

}
