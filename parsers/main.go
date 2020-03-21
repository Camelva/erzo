package parsers

import (
	"erzo/parsers/soundcloud"
	"erzo/parsers/youtube"
	"erzo/types"
	"fmt"
	"net/url"
)

func init() {
	soundcloudIE := soundcloud.IE
	youtubeIE := youtube.IE
	extractors = append(extractors, soundcloudIE, youtubeIE)
}

var extractors []types.Extractor

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
