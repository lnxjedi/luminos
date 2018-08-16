package host

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/bradleypeabody/fulltext"
	"github.com/lnxjedi/luminos/page"
)

func (host *Host) doSearch(w http.ResponseWriter, req *http.Request) {

	// Absolute document root.
	var docroot string
	var err error

	if docroot, err = host.GetContentPath(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var c bytes.Buffer
	if s, err := fulltext.NewSearcher(path.Join(host.DocumentRoot, "search.cdb")); err != nil {
		http.Error(w, fmt.Sprintf("unable to open search index: %v", err), http.StatusInternalServerError)
		return
	} else {
		defer s.Close()
		q := req.URL.Query()
		terms, ok := q["terms"]
		if ok {
			search := strings.Join(terms, " ")
			sr, err := s.SimpleSearch(search, 20)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed search: %v", err), http.StatusInternalServerError)
				return
			}
			if len(sr.Items) > 0 {
				c.Write([]byte(fmt.Sprintf("<h2>Search Results: \"%s\"</h2>", search)))
				c.Write([]byte("<ul>\n"))
				for _, v := range sr.Items {
					si := fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", v.StoreValue, v.StoreValue)
					c.Write([]byte(si))
				}
				c.Write([]byte("</ul>\n"))
			} else {
				c.Write([]byte("<h3>No results</h3>"))
			}
		} else {
			c.Write([]byte("<h3>No results</h3>"))
		}
	}

	p := &page.Page{}

	p.FilePath = docroot
	p.FileDir = docroot + pathSeparator
	p.BaseDir = ""
	p.BasePath = pathSeparator
	p.Title = "Search"
	p.IsHome = true

	p.Content = template.HTML(c.Bytes())

	// werc-like header and footer.
	hfile, hstat := guessFile(p.FileDir+"_header", true)

	if hstat != nil {
		hcontent, herr := host.readFile(hfile)
		if herr == nil {
			p.ContentHeader = template.HTML(hcontent)
		}
	}

	// werc-like header and footer.
	ffile, fstat := guessFile(p.FileDir+"_footer", true)

	if fstat != nil {
		fcontent, ferr := host.readFile(ffile)
		if ferr == nil {
			p.ContentFooter = template.HTML(fcontent)
		}
	}

	p.CreateBreadCrumb()
	p.CreateMenu()
	p.CreateSideMenu()

	p.ProcessContent()

	// Applying template.
	if err = host.Templates["index.tpl"].Execute(w, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}
