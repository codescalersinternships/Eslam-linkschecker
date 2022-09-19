package main

import (
	"reflect"

	toml "github.com/pelletier/go-toml"
)

func LinksFromConfig(configFile *string) ([]string, error) {
	config, err := toml.LoadFile(*configFile)

	if err != nil {
		return nil, err
	}
	mp := config.ToMap()
	return getLinksFromMap(mp), nil
}

func getLinksFromMap(mp any) []string {
	links := []string{}
	for _, val := range mp.(map[string]any) {
		if reflect.TypeOf(val).Kind() == reflect.Map {
			links = append(links, getLinksFromMap(val)...)
		} else {
			links = append(links, val.(string))
		}
	}
	return links
}
