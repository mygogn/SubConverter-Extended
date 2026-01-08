package config

import (
	"log"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type FetchFunc func(path string) (string, error)

func ImportItems(items []string, scopeLimit bool, fetch FetchFunc) ([]string, error) {
	var result []string
	for _, item := range items {
		if !utils.StartsWith(item, "!!import:") {
			result = append(result, item)
			continue
		}
		path := item[len("!!import:"):]
		var content string
		if utils.FileExists(path, scopeLimit) {
			content = utils.FileGet(path, scopeLimit)
		} else if utils.IsLink(path) {
			if fetch == nil {
				log.Printf("import: no fetcher for %s", path)
				continue
			}
			data, err := fetch(path)
			if err != nil {
				log.Printf("import: fetch failed for %s: %v", path, err)
				continue
			}
			content = data
		} else {
			log.Printf("import: invalid path %s", path)
			continue
		}
		if content == "" {
			continue
		}
		lines := utils.Split(content, string(utils.GetLineBreak(content)))
		for _, line := range lines {
			line = utils.TrimWhitespace(line, true, true)
			if line == "" {
				continue
			}
			if utils.StartsWith(line, ";") || utils.StartsWith(line, "#") || utils.StartsWith(line, "//") {
				continue
			}
			result = append(result, line)
		}
	}
	return result, nil
}
