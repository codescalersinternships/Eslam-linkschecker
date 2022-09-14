package main

import (
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	toml "github.com/pelletier/go-toml"
)

func main() {
	configFile := flag.String("config", "", "")
	flag.Parse()

	config, err := toml.LoadFile(*configFile)

	if err != nil {
		panic(err)
	}

	mp := config.ToMap()
	brokenLinks := checkMap(mp)

	for _, elem := range brokenLinks {
		fmt.Println(elem)
	}

}

func checkMap(mp any) []string {
	brokenLinks := []string{}
	for _, val := range mp.(map[string]any) {

		if reflect.TypeOf(val).Kind() == reflect.Map {
			brokenLinks = append(brokenLinks, checkMap(val)...)

		} else {

			if !validLink(val.(string)) {
				brokenLinks = append(brokenLinks, val.(string))
			}
		}
	}
	return brokenLinks
}

func validLink(link string) bool {
	if !strings.Contains(link, ("http")) {
		link = "http://" + link
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
