package crawler

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var defaultCrawler *Crawler

var defaultResp *http.Response

var defaultSite Site

var parentURL *url.URL
var homepage *url.URL
var about *url.URL

var emptyPage *httptest.Server
var emptyPageURL *url.URL

// {
// 	0:
// 		1:
// 			1,
// 			2,
// }
var mock0 *httptest.Server
var mock0URL *url.URL
var mock01 *httptest.Server
var mock01URL *url.URL
var mock011 *httptest.Server
var mock011URL *url.URL
var mock012 *httptest.Server
var mock012URL *url.URL

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defaultCrawler = NewCrawler(context.Background(), CrawlerOpt{})

	defaultResp = &http.Response{
		Header: http.Header{
			"Content-Type": {"text/html"},
		},
		Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`
			<html><body>
				<a href="/1">1</a>
			</body></html>`))),
	}

	defaultSite = Site{}

	parentURL, _ = url.Parse("https://monzo.com")
	homepage, _ = url.Parse("https://monzo.com/")
	about, _ = url.Parse("https://monzo.com/about")

	emptyPage = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body></body></html>`))
	}))
	defer emptyPage.Close()
	emptyPageURL, _ = url.Parse(emptyPage.URL)

	mock012 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Last page</body></html>`))
	}))
	defer mock012.Close()
	mock012URL, _ = url.Parse(mock012.URL)

	mock011 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Last page</body></html>`))
	}))
	defer mock011.Close()
	mock011URL, _ = url.Parse(mock011.URL)

	mock01 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html><body>
				<a href="` + mock011.URL + `">011</a>
				<a href="` + mock012.URL + `">012</a>
			</body></html>`))
	}))
	defer mock01.Close()
	mock01URL, _ = url.Parse(mock01.URL)

	mock0 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html><body>
				<a href="` + mock01.URL + `">01</a>
			</body></html>`))
	}))
	defer mock0.Close()
	mock0URL, _ = url.Parse(mock0.URL)

	return m.Run()
}
