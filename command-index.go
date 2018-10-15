// Copyright (c) 2018 David Parsley
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
	"fmt"
	"os"

	"github.com/lnxjedi/cli"
)

// runCommand is the structure that provides instructions for the "luminos
// " subcommand.
type indexCommand struct {
}

// Execute indexes all the sites.
func (c *indexCommand) Execute() (err error) {
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

	return nil
}

func init() {
	// Describing the "index" subcommand.

	cli.Register("index", cli.Entry{
		Name:        "index",
		Description: "Generates search index(es) for Luminos sites.",
		Arguments:   []string{"c"},
		Command:     &indexCommand{},
	})

}
