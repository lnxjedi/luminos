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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bradleypeabody/fulltext"
)

// TODO: read extensions for indexing from site config
var fileExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".html": true,
}

var camelCaseRegex = regexp.MustCompile("([a-z][a-z])([A-Z][a-z])")

func index(idxfile, content string) error {
	var f *os.File
	var err error

	content = path.Clean(content)

	f, err = os.Create(idxfile)
	if err != nil {
		return fmt.Errorf("creating index file: %v", err)
	}
	// create new index with temp dir (usually "" is fine)
	idx, err := fulltext.NewIndexer("")
	if err != nil {
		return fmt.Errorf("creating fulltext.NewIndexer: %v", err)
	}
	defer idx.Close()

	// stop words won't be indexed as search terms
	idx.StopWordCheck = fulltext.EnglishStopWordChecker

	findex := func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		cpath := strings.TrimPrefix(fpath, content)
		ext := path.Ext(cpath)
		bare := strings.TrimSuffix(path.Base(cpath), ext)
		pagetitle := strings.Replace(bare, "-", " ", -1)
		pagetitle = strings.Replace(pagetitle, "_", " ", -1)
		// try to detect camel case and insert spaces
		pagetitle = camelCaseRegex.ReplaceAllString(pagetitle, "$1 $2")
		if fileExtensions[path.Ext(cpath)] {
			c, err := ioutil.ReadFile(fpath)
			if err != nil {
				fmt.Printf("Error reading '%s' for indexing: %v\n", cpath, err)
			} else {
				fmt.Printf("Indexing %s: %s\n", pagetitle, cpath)
			}
			ic := bytes.NewBuffer(c)
			// Make sure page title gets indexed
			ic.Write([]byte(`\n` + pagetitle))
			// Index the content + title
			cdoc := fulltext.IndexDoc{
				Id:         []byte("t:" + cpath),
				StoreValue: []byte(cpath),
				IndexValue: ic.Bytes(),
			}
			idx.AddDoc(cdoc)
		}
		return nil
	}

	err = filepath.Walk(content, findex)
	if err != nil {
		return fmt.Errorf("walking '%s': %v", content, err)
	}

	err = idx.FinalizeAndWrite(f)
	if err != nil {
		return fmt.Errorf("finalizing index: %v", err)
	}
	return nil
}
