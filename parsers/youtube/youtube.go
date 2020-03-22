package youtube

import (
	"fmt"
	"net/url"
	"regexp"

	"erzo/engine"
	"erzo/parsers"
)

type Extractor struct {
	urlPattern string
	apiURL     string
	baseURL    string
}

var IE Extractor

var debugInstance = "youtube"

func init() {
	IE = Extractor{
		urlPattern: `(?:www\.)?(?:youtube\.com|youtu.be)`,
		apiURL:     "https://api.soundcloud.com/",
		baseURL:    "https://youtube.com/",
	}
	engine.AddExtractor(IE)
}

func (ie Extractor) Compatible(u url.URL) bool {
	s := u.Hostname()
	ok, err := regexp.MatchString(IE.urlPattern, s)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("comparing url: %s", err))
		return false
	}
	return ok
}

func (ie Extractor) Extract(u url.URL) (*parsers.ExtractorInfo, error) {
	_ = u
	info := &parsers.ExtractorInfo{}
	engine.Log(debugInstance, fmt.Errorf("extracting YouTube url"))
	return info, nil
}
