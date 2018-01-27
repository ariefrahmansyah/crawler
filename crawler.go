package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/ariefrahmansyah/href"
	"github.com/asaskevich/govalidator"
	"github.com/prometheus/common/log"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var defaultHTTPClient = &http.Client{}
var defaultMaxDepth = 2

type CrawlerOpt struct {
	HTTPClient *http.Client
	MaxDepth   int
}

type Crawler struct {
	httpClient       *http.Client
	visitedSite      map[string]Site
	visitedSiteMutex *sync.Mutex
}

func NewCrawler(ctx context.Context, opt CrawlerOpt) *Crawler {
	crawler := &Crawler{
		httpClient:       defaultHTTPClient,
		visitedSite:      make(map[string]Site),
		visitedSiteMutex: &sync.Mutex{},
	}

	if opt.HTTPClient != nil {
		crawler.httpClient = opt.HTTPClient
	}

	return crawler
}

type CrawlQuery struct {
	Site     string `valid:"url,required"`
	MaxDepth int    `valid:"-"`
	Timeout  int    `valid:"-"`
}

func (crawler *Crawler) Crawl(ctx context.Context, query CrawlQuery, depth int) (Site, error) {
	if query.MaxDepth == 0 {
		query.MaxDepth = defaultMaxDepth
	}

	if depth >= query.MaxDepth {
		log.Debugf("Depth is bigger than threshold. Stopping.")
		return Site{}, nil
	}

	if valid, err := crawler.Validate(ctx, query); !valid && err != nil {
		return Site{}, fmt.Errorf("Query is not valid. { %v }", err)
	}

	siteURL, err := url.Parse(query.Site)
	if err != nil {
		return Site{}, fmt.Errorf("Failed to parse URL ( %s ). { %v }", query.Site, err)
	}
	log.Debugf("URL to be crawled: %s", siteURL)

	visited, err := crawler.GetSiteFromCache(ctx, siteURL)
	if err == nil {
		log.Debugf("Already visited. Fetch from cache ( %s )", siteURL)
		return visited, nil
	}

	resp, err := crawler.Fetch(ctx, siteURL)
	if err != nil {
		return Site{}, fmt.Errorf("Failed to fetch page ( %s ). { %v }", query.Site, err)
	}
	log.Debugf("Response ( %s ): %s", siteURL, resp.Status)

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if !crawler.IsWebpage(ctx, resp) {
		log.Debugf("Not a webpage. Do not crawl ( %s )", siteURL)
		return Site{}, nil
	}

	// Get links on the page
	links, err := crawler.GetLinks(ctx, siteURL, resp, depth+1)
	if err != nil {
		return Site{}, fmt.Errorf("Failed to get links ( %s ). { %v }", query.Site, err)
	}

	site := Site{
		mutex: &sync.Mutex{},
		Data:  href.NewLink(ctx, siteURL, "", siteURL.String(), depth),
	}

	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func(site *Site, link href.Link) {
			defer wg.Done()
			log.Infof("%s", link)

			linkQuery := CrawlQuery{
				Site:     link.URL.String(),
				MaxDepth: query.MaxDepth,
			}

			s, err := crawler.Crawl(ctx, linkQuery, depth+1)
			if err != nil {
				log.Errorf("Failed to crawl ( %s ). { %v }", link.URL, err)
				return
			}
			s.Data = link

			site.AppendSite(s)
		}(&site, link)
	}
	wg.Wait()

	sort.Sort(SitesSorter(site.Sites))

	crawler.PutSiteToCache(ctx, siteURL, site)

	return site, nil
}

func (crawler *Crawler) Validate(ctx context.Context, query CrawlQuery) (bool, error) {
	_, err := govalidator.ValidateStruct(query)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (crawler *Crawler) Fetch(ctx context.Context, siteURL *url.URL) (*http.Response, error) {
	if siteURL.Scheme == "" {
		siteURL.Scheme = "http"
	}

	resp, err := crawler.httpClient.Get(siteURL.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to get page (%s). { %v }", siteURL, err)
	}

	if resp.StatusCode > http.StatusOK {
		return nil, fmt.Errorf("Failed to get page (%s). { Response status code = %d }", siteURL, resp.StatusCode)
	}

	return resp, nil
}

func (crawler *Crawler) GetSiteFromCache(ctx context.Context, siteURL *url.URL) (Site, error) {
	crawler.visitedSiteMutex.Lock()
	defer crawler.visitedSiteMutex.Unlock()

	visited, ok := crawler.visitedSite[siteURL.String()]
	if ok {
		return visited, nil
	}

	return Site{}, fmt.Errorf("Page is not cached yet ( %s )", siteURL)
}

func (crawler *Crawler) PutSiteToCache(ctx context.Context, siteURL *url.URL, site Site) error {
	crawler.visitedSiteMutex.Lock()
	defer crawler.visitedSiteMutex.Unlock()

	crawler.visitedSite[siteURL.String()] = site

	return nil
}

func (crawler Crawler) IsWebpage(ctx context.Context, resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return contentType == "text/html" || strings.Contains(contentType, "text/html")
}

func (crawler Crawler) GetLinks(ctx context.Context, siteURL *url.URL, resp *http.Response, depth int) (map[string]href.Link, error) {
	links := make(map[string]href.Link)

	// Parse the page.
	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Search for anchors tag.
	anchors := scrape.FindAll(root, scrape.ByTag(atom.A))

	for _, anchor := range anchors {
		aText := scrape.Text(anchor)
		if aText == "" {
			continue
		}

		aHref := scrape.Attr(anchor, "href")
		if aHref == "" {
			continue
		}

		link := href.NewLink(ctx, siteURL, aText, aHref, depth)

		if link.IsValidPageLink(ctx) {
			if href.IsSameDomain(siteURL, link.URL) {
				log.Debugf("Link to be crawled: %s", link.URL)
				links[link.URL.String()] = link
			} else {
				log.Debugf("Out of domain. Do not crawl: %s", link.HREF)
			}
		}
	}

	return links, nil
}
