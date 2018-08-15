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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bradleypeabody/fulltext"
)

// TODO: read stop words and extensions from site settings
var localStopWords = map[string]bool{
	"README":    true,
	"index":     true,
	"CHANGELOG": true,
}

var fileExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".html": true,
}

func index(docroot, content string) error {
	var f *os.File
	var err error

	content = path.Clean(content)

	f, err = os.Create(path.Join(docroot, "index.cdb"))
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
	idx.StopWordCheck = func(s string) bool {
		if localStopWords[s] {
			return true
		}
		return fulltext.STOPWORDS_EN[s]
	}

	findex := func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		cpath := strings.TrimPrefix(fpath, content)
		ext := path.Ext(cpath)
		bare := strings.TrimSuffix(path.Base(cpath), ext)
		if fileExtensions[path.Ext(cpath)] {
			c, err := ioutil.ReadFile(fpath)
			if err != nil {
				fmt.Printf("Error reading '%s' for indexing: %v\n", cpath, err)
			} else {
				fmt.Printf("Indexing %s: %s\n", bare, cpath)
			}
			// Index the title
			tdoc := fulltext.IndexDoc{
				Id:         []byte("t:" + cpath),
				StoreValue: []byte(cpath),
				IndexValue: []byte(bare),
			}
			idx.AddDoc(tdoc)
			// Index the content
			cdoc := fulltext.IndexDoc{
				Id:         []byte(cpath),
				StoreValue: []byte(cpath),
				IndexValue: c,
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
