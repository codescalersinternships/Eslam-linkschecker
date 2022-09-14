package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	toml "github.com/pelletier/go-toml"
	"golang.org/x/net/html"
)

func main() {
	configFile := flag.String("config", "config.toml", "")
	flag.Parse()

	config, err := toml.LoadFile(*configFile)

	if err != nil {
		panic(err)
	}

	mp := config.ToMap()

	checkMap(mp)

}

func checkMap(mp any) {
	for _, val := range mp.(map[string]any) {

		if reflect.TypeOf(val).Kind() == reflect.Map {
			checkMap(val)
		} else {
			ch := make(chan bool)
			go validateWabsite(val.(string), ch)
			<-ch
		}
	}
}

func validateWabsite(host string, ch chan bool) {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = fmt.Sprintf("http://%s", host)
	}
	resp, err := http.Get(host)

	if err != nil {
		fmt.Println(host)
		return
	}

	defer resp.Body.Close()

	links := getLinks(resp.Body)

	for _, link := range links {
		if !validateLink(link) {
			fmt.Println("failed: ", link)
		} else {
			fmt.Println("passed: ", link)
		}
	}
	ch <- true
}

func validateLink(link string) bool {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = fmt.Sprintf("http://%s", link)
	}

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

func getLinks(body io.Reader) []string {
	var links []string
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
