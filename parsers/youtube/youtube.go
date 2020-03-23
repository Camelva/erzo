package youtube

import (
	"net/url"
	"regexp"

	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
)

type Extractor struct {
	urlPattern string
	apiURL     string
	baseURL    string
}

var IE Extractor

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
	ok, _ := regexp.MatchString(IE.urlPattern, s)
	return ok
}

func (ie Extractor) Extract(u url.URL) (*parsers.ExtractorInfo, error) {
	_ = u
	info := &parsers.ExtractorInfo{}
	return info, parsers.ErrFormatNotSupported{Format: "Youtube"}
}
