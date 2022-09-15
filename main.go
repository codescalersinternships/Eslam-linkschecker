package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	configFile := flag.String("config", "config.toml", "")
	flag.Parse()

	links := parseConfigFile(configFile)

	mp := make(map[string]bool)
	checkAllLinks(links, &mp, "")
}

func validLink(link string) bool {
	link = addHeader(link)

	headResp, headErr := http.Head(link)
	getResp, GetErr := http.Get(link)

	if headErr != nil || GetErr != nil {
		return false
	}

	defer headResp.Body.Close()
	defer getResp.Body.Close()

	getStatusCode := getResp.StatusCode
	headStatusCode := headResp.StatusCode
	return (getStatusCode >= 200 && getStatusCode < 400) ||
		(headStatusCode >= 200 && headStatusCode < 400)
}

func getLinks(link string) []string {
	var links []string
	link = addHeader(link)

	resp, err := http.Get(link)

	if err != nil {
		return nil
	}

	body := resp.Body
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

func checkAllLinks(links []string, allLinks *map[string]bool, parent string) {
	for _, link := range links {
		if !(*allLinks)[link] {
			(*allLinks)[link] = true
			if validLink(link) {
				innerLinks := getLinks(link)
				checkAllLinks(innerLinks, allLinks, link)
			} else if parent != "" {
				host := getHostname(parent)
				newLink := fmt.Sprintf("%s/%s", host, strings.Trim(link, "/"))
				if host != "" && validLink(newLink) {
					innerLinks := getLinks(addHeader(newLink))
					checkAllLinks(innerLinks, allLinks, newLink)
				} else {
					fmt.Println("Failed: ", newLink)
				}
			} else {
				fmt.Println("Failed: ", link)
			}
		}
	}
}

func addHeader(link string) string {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = fmt.Sprintf("http://%s", link)
	}
	return link
}

func getHostname(link string) string {
	link = addHeader(link)
	url, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return strings.Trim(url.Hostname(), "/")
}
