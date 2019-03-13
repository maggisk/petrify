package main

import "os"
import "path/filepath"
import "strings"
import "regexp"
import "bytes"
import "golang.org/x/net/html"
import "net/http"
import urlpkg "net/url"
import "io/ioutil"

type Crawler struct {
	visited   map[string]bool
	root      string
	buildDir  string
	linkRules map[string]bool
}

func NewCrawler(root string, buildDir string, linkRules string) *Crawler {
	crawler := &Crawler{
		visited:   make(map[string]bool),
		root:      root,
		buildDir:  buildDir,
		linkRules: make(map[string]bool),
	}

	// create a lookup table from link rule string
	for _, rule := range strings.Split(strings.ToLower(linkRules), " ") {
		crawler.linkRules[rule] = true
	}

	return crawler
}

func (crawler *Crawler) CrawlAll(urls []string) {
	for _, url := range urls {
		crawler.Crawl(url)
	}
}

func (crawler *Crawler) Crawl(path string) {
	url := crawler.root + path

	if crawler.visited[url] {
		return
	}
	crawler.visited[url] = true

	info("Crawling '%s'", url)

	// fetch page
	resp, err := http.Get(url)
	checkError(err)

	if resp.StatusCode != 200 {
		warn("'%s' responded with status '%s'", url, resp.Status)
	}

	// write page to file
	contentType := getContentType(resp)
	filePath := crawler.buildDir + urlToFilePath(url, contentType)
	crawler.SaveFile(resp, filePath)

	// follow internal links in html pages
	if strings.HasPrefix(contentType, "text/html") {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		for _, link := range extractLinks(body, crawler.linkRules) {
			crawler.Crawl(link)
		}
	}
}

func (crawler *Crawler) SaveFile(resp *http.Response, filePath string) {
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	checkError(os.MkdirAll(filepath.Dir(filePath), DIR_BITMASK))
	checkError(ioutil.WriteFile(filePath, body, FILE_BITMASK))
}

func extractLinks(body []byte, linkRules map[string]bool) []string {
	doc, err := html.Parse(bytes.NewReader(body))
	checkError(err)

	links := []string{}
	iterateAllNodes(doc, func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if shouldCheckLink(linkRules, node.Data, attr.Key) && isRelativeURL(attr.Val) {
					links = append(links, attr.Val)
				}
			}
		}
	})

	return links
}

func iterateAllNodes(node *html.Node, f func(*html.Node)) {
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		f(node)
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(node)
}

func isRelativeURL(url string) bool {
	// TODO: see if we can use urlpkg.Parse() to do this more reliably

	// protocol prefix, http: mailto: javascript: etc
	if regexp.MustCompile(`^\w+:`).MatchString(url) {
		return false
	}

	// protocol-relative url
	if strings.HasPrefix(url, "//") {
		return false
	}

	// otherwise assume url is within our site
	return true
}

func getContentType(resp *http.Response) string {
	if len(resp.Header["Content-Type"]) == 0 {
		return "text/html"
	}
	return resp.Header["Content-Type"][0]
}

func urlToFilePath(url string, contentType string) string {
	u, err := urlpkg.Parse(url)
	checkError(err)
	path := u.Path

	// rewrite html paths from e.g. /about or /about/ to /about/index.html
	if strings.HasPrefix(contentType, "text/html") && !strings.HasSuffix(path, ".html") {
		if !strings.HasSuffix(path, "/") {
			path = path + "/"
		}
		path = path + "index.html"
	}

	return filepath.FromSlash(path)
}

func shouldCheckLink(rules map[string]bool, tag, attr string) bool {
	return rules[tag+"."+attr] || rules[tag+".*"] || rules["*."+attr] || rules["*.*"]
}
