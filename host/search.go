package host

import (
	"log"
	"strings"

	"github.com/bradleypeabody/fulltext"
)

// Search returns the results of a fulltext search on terms.
func (host *Host) Search(terms []string, res int) []fulltext.SearchResultItem {

	empty := make([]fulltext.SearchResultItem, 0)

	if len(terms) == 0 || res == 0 {
		return empty
	}

	ifile, ierr := host.GetIndexPath()
	if ierr != nil {
		log.Printf("getting index path for host %s: %v", host.Name, ierr)
		return empty
	}

	s, err := fulltext.NewSearcher(ifile)
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
