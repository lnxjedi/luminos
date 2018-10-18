package host

import (
	"log"
	"path"
	"strings"

	"github.com/bradleypeabody/fulltext"
)

// Search returns the results of a fulltext search on terms.
func (host *Host) Search(terms []string, res int) []fulltext.SearchResultItem {

	empty := make([]fulltext.SearchResultItem, 0)

	if len(terms) == 0 || res == 0 {
		return empty
	}

	s, err := fulltext.NewSearcher(path.Join(host.DocumentRoot, "search.cdb"))
	if err != nil {
		log.Printf("unable to open search index for host %s: %v", host.Name, err)
		return empty
	}
	defer s.Close()
	search := strings.Join(terms, " ")
	sr, err := s.SimpleSearch(search, res)
	if err != nil {
		log.Printf("failed search: %v", err)
		return empty
	}
	return sr.Items
}

/* 	if len(sr.Items) > 0 {
		l := make(map[string]struct{})
		c.Write([]byte(fmt.Sprintf("<h2>Search Results: \"%s\"</h2>", search)))
		c.Write([]byte("<ul>\n"))
		for _, v := range sr.Items {
			key := string(v.StoreValue)
			_, found := l[key]
			if !found {
				si := fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", v.StoreValue, v.StoreValue)
				c.Write([]byte(si))
				l[key] = struct{}{}
			}
		}
		c.Write([]byte("</ul>\n"))
	} else {
		c.Write([]byte("<h3>No results</h3>"))
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
		hcontent, herr := host.readContentFile(hfile)
		if herr == nil {
			p.ContentHeader = template.HTML(hcontent.Content)
		}
	}

	// werc-like header and footer.
	ffile, fstat := guessFile(p.FileDir+"_footer", true)

	if fstat != nil {
		fcontent, ferr := host.readContentFile(ffile)
		if ferr == nil {
			p.ContentFooter = template.HTML(fcontent.Content)
		}
	}

	p.CreateBreadCrumb()
	p.CreateMenu()
	p.CreateSideMenu()

	// p.GenerateTitles()

	host.RLock()
	ht := host.TemplateGroup
	host.RUnlock()
	// Applying template.
	if err = ht.ExecuteTemplate(w, "index.tpl", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
} */
