package crawler

import (
	"sort"
	"sync"

	"github.com/ariefrahmansyah/href"
)

// Site struct.
type Site struct {
	mutex *sync.Mutex
	Data  href.Link `json:"data"`
	Sites []Site    `json:"site,omitempty"`
}

// AppendSite add sitemap to site.
func (site *Site) AppendSite(s Site) {
	site.mutex.Lock()
	defer site.mutex.Unlock()

	site.Sites = append(site.Sites, s)
}

// SitesSorter sorts sites by text.
type SitesSorter []Site

func (sites SitesSorter) Len() int           { return len(sites) }
func (sites SitesSorter) Swap(i, j int)      { sites[i], sites[j] = sites[j], sites[i] }
func (sites SitesSorter) Less(i, j int) bool { return sites[i].Data.Text < sites[j].Data.Text }

// SortSites sorts sites by text ascending.
func SortSites(sites []Site) {
	sort.Sort(SitesSorter(sites))
}
