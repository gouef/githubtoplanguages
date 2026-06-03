package requests

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type LinguistLanguage struct {
	Color      string   `yaml:"color"`
	Extensions []string `yaml:"extensions"`
}

func LoadLinguistLanguages() (map[string]struct{ Name, Color string }, error) {
	url := "https://raw.githubusercontent.com/github-linguist/linguist/master/lib/linguist/languages.yml"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download linguist languages, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawData map[string]LinguistLanguage
	if err := yaml.Unmarshal(body, &rawData); err != nil {
		return nil, err
	}

	extensionMap := make(map[string]struct{ Name, Color string })
	for langName, langInfo := range rawData {
		color := langInfo.Color
		if color == "" {
			color = "#cccccc"
		}

		for _, ext := range langInfo.Extensions {
			extensionMap[strings.ToLower(ext)] = struct{ Name, Color string }{
				Name:  langName,
				Color: color,
			}
		}
	}

	return extensionMap, nil
}
