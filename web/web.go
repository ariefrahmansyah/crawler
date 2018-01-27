package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/ariefrahmansyah/crawler"
	promnegroni "github.com/ariefrahmansyah/negroni-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

func main() {
	// log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", prometheus.Handler())

	// crawl a web page
	mux.HandleFunc("/crawl", CrawlHandler)

	promMiddleware := promnegroni.NewPromMiddleware("crawler", promnegroni.PromMiddlewareOpts{})

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.Use(promMiddleware)
	n.UseHandler(mux)

	log.Infof("App started at port: %s", port)
	http.ListenAndServe(":"+port, n)
}

func CrawlHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.ParseForm()

	site := r.FormValue("site")
	maxDepthStr := r.FormValue("max_depth")
	maxDepth, _ := strconv.Atoi(maxDepthStr)

	crawlQuery := crawler.CrawlQuery{
		Site:     site,
		MaxDepth: maxDepth,
	}

	crawl := crawler.NewCrawler(ctx, crawler.CrawlerOpt{})
	sitemap, err := crawl.Crawl(ctx, crawlQuery, 0)
	if err != nil {
		log.Errorf("Failed to crawl ( %v ). { %s }", crawlQuery, err)
		w.Write([]byte(err.Error()))
		return
	}

	sitemapJSON, err := json.Marshal(sitemap)
	if err != nil {
		log.Errorf("Failed to marshal sitemap ( %v ). { %s }", sitemap, err)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(sitemapJSON)
}
