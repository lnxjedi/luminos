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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
	"github.com/lnxjedi/dig"
	"github.com/lnxjedi/luminos/host"
	"github.com/lnxjedi/to"
)

// Map of hosts.
var hosts map[string]*host.Host

// File watcher.
var watch *fsnotify.Watcher

type server struct {
}

func init() {
	// Allocating map.
	hosts = make(map[string]*host.Host)
}

// Finds the appropriate hosts for a request.
func route(req *http.Request) *host.Host {

	// Request's hostname.
	name := req.Host

	// Removing the port part of the host.
	if strings.Contains(name, ":") {
		name = name[0:strings.Index(name, ":")]
	}

	// Host and path.
	path := name + req.URL.Path

	// Searching for the host that best matches this request.
	match := ""

	for key := range hosts {
		lkey := len(key)
		if key[0] == '/' {
			if lkey <= len(req.URL.Path) {
				if req.URL.Path[0:lkey] == key {
					match = key
				}
			}
		}
		if lkey <= len(path) {
			if path[0:lkey] == key {
				match = key
			}
		}
	}

	// No host matched, let's use the default host.
	if match == "" {
		// log.Printf("Path %v could not match any route, falling back to the default.\n", path)
		match = "default"
	}

	// Let's verify and return the host.
	if _, ok := hosts[match]; !ok {
		// Host was not found.
		log.Printf("Request for unknown host: %s\n", req.Host)
		return nil
	}

	return hosts[match]
}

// Routes a request and lets the host handle it.
func (s server) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	r := route(req)
	if r != nil {
		r.ServeHTTP(wri, req)
	} else {
		log.Printf("Failed to serve host %s.\n", req.Host)
		http.Error(wri, "Not found", http.StatusNotFound)
	}
}

// Loads settings
func loadSettings() (dig.InterfaceMap, error) {

	var entries map[string]interface{}
	var ok bool

	// Trying to read settings from file.
	_, err := os.Stat(*flagSettings)
	var y dig.InterfaceMap
	if err == nil {
		if ydata, err := ioutil.ReadFile(*flagSettings); err != nil {
			return nil, fmt.Errorf(`could not read settings file (%s): %q`, *flagSettings, err)
		} else {
			if err := yaml.Unmarshal(ydata, &y); err != nil {
				return nil, fmt.Errorf(`parsing settings file (%s): %q`, *flagSettings, err)
			}
		}
	} else {
		return nil, fmt.Errorf(`error trying to open settings file (%s): %q`, *flagSettings, err)
	}

	// Loading and verifying host entries
	e := y.Get("hosts")
	if entries, ok = e.(map[string]interface{}); !ok {
		return nil, errors.New("missing 'hosts' entry")
	}

	h := map[string]*host.Host{}

	// Populating host entries.
	for name := range entries {
		path := to.String(entries[name])

		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to validate host %s: %q", name, err)
		}
		if info.IsDir() == false {
			return nil, fmt.Errorf("host %s does not point to a directory", name)
		}

		h[name], err = host.New(name, path)

		if err != nil {
			return nil, fmt.Errorf("failed to initialize host %s: %q", name, err)
		}
	}

	for name := range hosts {
		hosts[name].Close()
	}

	hosts = h

	if _, ok := hosts["default"]; ok == false {
		log.Printf("Warning: default host was not provided.\n")
	}

	return y, nil
}

func settingsWatcher() (*fsnotify.Watcher, error) {

	var err error

	watcher, err := fsnotify.NewWatcher()

	if err == nil {

		go func() {
			defer watcher.Close()
			for {
				select {
				case ev, ok := <-watcher.Events:
					if !ok {
						return
					}
					log.Printf("luminos got ev: %v\n", ev)

					y, err := loadSettings()
					if err != nil {
						log.Printf("Error loading settings file %s: %q\n", ev.Name, err)
					} else {
						settings = y
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Printf("Watcher error: %q\n", err)
				}
			}
		}()
	}

	return watcher, err
}
