package crawler

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"

	"github.com/ariefrahmansyah/href"
)

func TestNewCrawler(t *testing.T) {
	type args struct {
		ctx context.Context
		opt CrawlerOpt
	}
	tests := []struct {
		name string
		args args
		want *Crawler
	}{
		{
			"custom http client",
			args{
				context.Background(),
				CrawlerOpt{
					HTTPClient: &http.Client{},
				},
			},
			&Crawler{
				httpClient:       &http.Client{},
				visitedSite:      make(map[string]Site),
				visitedSiteMutex: &sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCrawler(tt.args.ctx, tt.args.opt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCrawler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_Crawl(t *testing.T) {
	type args struct {
		ctx   context.Context
		query CrawlQuery
		depth int
	}
	tests := []struct {
		name    string
		crawler *Crawler
		args    args
		want    Site
		wantErr bool
	}{
		{
			"depth exceeded",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					MaxDepth: 1,
				},
				2,
			},
			Site{},
			false,
		},
		{
			"invalid query's site",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					Site: "invalid site",
				},
				0,
			},
			Site{},
			true,
		},
		{
			"fetch empty page",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					Site: emptyPage.URL,
				},
				0,
			},
			Site{
				mutex: &sync.Mutex{},
				Data:  href.NewLink(context.Background(), emptyPageURL, "", emptyPageURL.String(), 0),
			},
			false,
		},
		{
			"crawl mock 0",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					Site: mock0.URL,
				},
				0,
			},
			Site{
				mutex: &sync.Mutex{},
				Data:  href.NewLink(context.Background(), mock0URL, "", mock0URL.String(), 0),
				Sites: []Site{
					Site{
						mutex: &sync.Mutex{},
						Data:  href.NewLink(context.Background(), mock01URL, "01", mock01URL.String(), 1),
						Sites: []Site{
							Site{
								Data:  href.NewLink(context.Background(), mock011URL, "011", mock011URL.String(), 2),
								Sites: nil,
							},
							Site{
								Data:  href.NewLink(context.Background(), mock012URL, "012", mock012URL.String(), 2),
								Sites: nil,
							},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.crawler.Crawl(tt.args.ctx, tt.args.query, tt.args.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawler.Crawl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawler.Crawl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_Validate(t *testing.T) {
	type args struct {
		ctx   context.Context
		query CrawlQuery
	}
	tests := []struct {
		name    string
		crawler *Crawler
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					Site: "https://monzo.com",
				},
			},
			true,
			false,
		},
		{
			"failed",
			defaultCrawler,
			args{
				context.Background(),
				CrawlQuery{
					Site: "monzo",
				},
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.crawler.Validate(tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawler.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Crawler.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_Fetch(t *testing.T) {
	type args struct {
		ctx     context.Context
		siteURL *url.URL
	}
	tests := []struct {
		name    string
		crawler *Crawler
		args    args
		want    *http.Response
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.crawler.Fetch(tt.args.ctx, tt.args.siteURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawler.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawler.Fetch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_GetSiteFromCache(t *testing.T) {
	type args struct {
		ctx     context.Context
		siteURL *url.URL
	}
	tests := []struct {
		name    string
		crawler *Crawler
		args    args
		want    Site
		wantErr bool
	}{
		{
			"page not cached",
			&Crawler{
				visitedSite:      make(map[string]Site),
				visitedSiteMutex: &sync.Mutex{},
			},
			args{
				context.Background(),
				parentURL,
			},
			Site{},
			true,
		},
		{
			"page cached",
			&Crawler{
				visitedSite: map[string]Site{
					parentURL.String(): defaultSite,
				},
				visitedSiteMutex: &sync.Mutex{},
			},
			args{
				context.Background(),
				parentURL,
			},
			defaultSite,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.crawler.GetSiteFromCache(tt.args.ctx, tt.args.siteURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawler.GetSiteFromCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawler.GetSiteFromCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_PutSiteToCache(t *testing.T) {
	type args struct {
		ctx     context.Context
		siteURL *url.URL
		site    Site
	}
	tests := []struct {
		name    string
		crawler *Crawler
		args    args
		wantErr bool
	}{
		{
			"1",
			&Crawler{
				visitedSite:      make(map[string]Site),
				visitedSiteMutex: &sync.Mutex{},
			},
			args{
				context.Background(),
				parentURL,
				defaultSite,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.crawler.PutSiteToCache(tt.args.ctx, tt.args.siteURL, tt.args.site); (err != nil) != tt.wantErr {
				t.Errorf("Crawler.PutSiteToCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrawler_IsWebpage(t *testing.T) {
	type args struct {
		ctx  context.Context
		resp *http.Response
	}
	tests := []struct {
		name    string
		crawler Crawler
		args    args
		want    bool
	}{
		{
			"a webpage",
			*defaultCrawler,
			args{
				context.Background(),
				defaultResp,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.crawler.IsWebpage(tt.args.ctx, tt.args.resp); got != tt.want {
				t.Errorf("Crawler.IsWebpage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawler_GetLinks(t *testing.T) {
	type args struct {
		ctx     context.Context
		siteURL *url.URL
		resp    *http.Response
		depth   int
	}
	tests := []struct {
		name    string
		crawler Crawler
		args    args
		want    map[string]href.Link
		wantErr bool
	}{
		{
			"1",
			*defaultCrawler,
			args{
				context.Background(),
				parentURL,
				defaultResp,
				0,
			},
			map[string]href.Link{
				"https://monzo.com/1": href.NewLink(context.Background(), parentURL, "1", "/1", 0),
			},
			false,
		},
		{
			"2",
			*defaultCrawler,
			args{
				context.Background(),
				parentURL,
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`
						<html><body>
							<a href="/1">1</a>
							<a href="https://monzo.com/2">2</a>
						</body></html>`))),
				},
				0,
			},
			map[string]href.Link{
				"https://monzo.com/1": href.NewLink(context.Background(), parentURL, "1", "/1", 0),
				"https://monzo.com/2": href.NewLink(context.Background(), parentURL, "2", "https://monzo.com/2", 0),
			},
			false,
		},
		{
			"no text in href #2",
			*defaultCrawler,
			args{
				context.Background(),
				parentURL,
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`
						<html><body>
							<a href="/1">1</a>
							<a href="https://monzo.com/2"></a>
						</body></html>`))),
				},
				0,
			},
			map[string]href.Link{
				"https://monzo.com/1": href.NewLink(context.Background(), parentURL, "1", "/1", 0),
			},
			false,
		},
		{
			"no href in #2",
			*defaultCrawler,
			args{
				context.Background(),
				parentURL,
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`
						<html><body>
							<a href="/1">1</a>
							<a href="">2</a>
						</body></html>`))),
				},
				0,
			},
			map[string]href.Link{
				"https://monzo.com/1": href.NewLink(context.Background(), parentURL, "1", "/1", 0),
			},
			false,
		},
		{
			"out of domain",
			*defaultCrawler,
			args{
				context.Background(),
				parentURL,
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`
						<html><body>
							<a href="/1">1</a>
							<a href="https://monzo.com/2">2</a>
							<a href="https://mondo.com">mondo</a>
						</body></html>`))),
				},
				0,
			},
			map[string]href.Link{
				"https://monzo.com/1": href.NewLink(context.Background(), parentURL, "1", "/1", 0),
				"https://monzo.com/2": href.NewLink(context.Background(), parentURL, "2", "https://monzo.com/2", 0),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.crawler.GetLinks(tt.args.ctx, tt.args.siteURL, tt.args.resp, tt.args.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("Crawler.GetLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Crawler.GetLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}
