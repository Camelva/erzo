package parsers

import (
	"erzo/parsers/soundcloud"
	"erzo/types"
	"fmt"
	"net/url"
)

type Extractor interface {
	Extract(url.URL) (*types.ExtractorInfo, error)
	Compatible(string) bool
}

var extractors []Extractor

func init() {
	soundcloudIE := soundcloud.Init()
	extractors = append(extractors, soundcloudIE)
}

func Parse(u url.URL) (*types.ExtractorInfo, error) {
	for _, extractor := range extractors {
		if !extractor.Compatible(u.Hostname()) {
			continue
		}
		res, err := extractor.Extract(u)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, fmt.Errorf("there is no compatible extractor")
}
