package youtube

import (
	"log"
	"net/url"
	"regexp"

	"erzo/types"
)

var IE Extractor

func init() {
	IE = Extractor{
		urlPattern: `(?:www\.)?(?:youtube\.com|youtu.be)`,
		apiURL:     "https://api.soundcloud.com/",
		baseURL:    "https://youtube.com/",
	}
	return
}

func (ie Extractor) Compatible(s string) bool {
	ok, err := regexp.MatchString(IE.urlPattern, s)
	if err != nil {
		log.Printf("[soundcloud] comparing: %s\n", err)
		return false
	}
	return ok
}

func (ie Extractor) Extract(u url.URL) (*types.ExtractorInfo, error) {
	_ = u
	info := &types.ExtractorInfo{}
	log.Println("Extracting youtube link")
	return info, nil
}
