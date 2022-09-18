package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	configFile := flag.String("config", "config.toml", "")
	flag.Parse()

	links, err := LinksFromConfig(configFile)

	if err != nil {
		panic(err)
	}

	mp := make(map[string]bool)
	checkAllLinks(links, mp, "")

}

func checkAllLinks(links []string, visitedLinks map[string]bool, parent string) {
	for _, link := range links {
		if !visitedLinks[link] {

			visitedLinks[link] = true
			tempLink := fmt.Sprintf("%s/%s", getHostname(parent), strings.Trim(link, "/"))

			if validLink(tempLink) {

				innerLinks := getLinks(tempLink)
				checkAllLinks(innerLinks, visitedLinks, tempLink)

			} else {

				if getHostname(link) == getHostname(parent) || parent == "" {

					if validLink(link) {
						innerLinks := getLinks(link)
						checkAllLinks(innerLinks, visitedLinks, link)
					} else {
						fmt.Println(link)
					}

				} else if !validLink(link) {
					fmt.Println(link)
				}
			}
		}
	}
}

func validLink(link string) bool {
	link = ensureScheme(link)
	var requestFun func(fn func(string) (*http.Response, error)) bool

	requestFun = func(fn func(string) (*http.Response, error)) bool {
		resp, err := fn(link)
		if err != nil {
			return false
		}

		defer resp.Body.Close()

		statusCode := resp.StatusCode
		return (statusCode >= 200 && statusCode < 400)
	}
	return requestFun(http.Head) || requestFun(http.Get)
}

func extractLinksFromString(link string) io.ReadCloser {
	link = ensureScheme(link)
	resp, err := http.Get(link)

	if err != nil {
		return nil
	}

	return resp.Body
}

func getLinks(link string) []string {
	var links []string

	body := extractLinksFromString(link)
	defer body.Close()

	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func ensureScheme(link string) string {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = fmt.Sprintf("https://%s", link)
	}
	return link
}

func getHostname(link string) string {
	link = ensureScheme(link)
	url, err := url.Parse(link)

	if err != nil {
		return ""
	}

	return strings.Trim(url.Hostname(), "/")
}
