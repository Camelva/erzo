package youtube

import (
	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
	"net/http"
	"net/url"
	"regexp"
)

type Extractor struct {
	name       string
	urlPattern string
}

func init() {
	var IE = Extractor{
		urlPattern: `(?:www\.)?(?:youtube\.com|youtu.be)`,
		name:       "Youtube",
	}
	engine.AddExtractor(IE)
}

func (ie Extractor) Name() string {
	return ie.name
}

func (ie Extractor) Compatible(u *url.URL) bool {
	s := u.Hostname()
	ok, _ := regexp.MatchString(ie.urlPattern, s)
	return ok
}

var audioITags = []int{251, 250, 249, 172, 171, 328, 325, 258, 256, 141, 140, 139}

func (ie Extractor) Extract(u *url.URL, debug bool, client *http.Client) (*parsers.ExtractorInfo, error) {
	c := &Client{HTTPClient: client, Debug: debug}
	video, err := c.GetVideo(u.String())
	if err != nil {
		return nil, err
	}
	formats := parsers.Formats{}

	if len(video.Streams) < 1 {
		return nil, parsers.ErrCantContinue("found no streams")
	}

	// add first available iTag for emergency usage
	audioITags = append(audioITags, video.Streams[0].ItagNo)

	for _, tag := range audioITags {
		if len(formats) > 2 {
			break
		}
		stream := video.FindStreamByItag(tag)
		if stream == nil {
			continue
		}
		uri, err := c.GetStreamURL(video, stream)
		if err != nil {
			continue
		}

		f := parsers.Format{
			Url:      uri,
			Ext:      "",
			Type:     "",
			Protocol: "https",
			Score:    0,
		}
		formats = append(formats, f)
	}

	info := parsers.ExtractorInfo{
		Permalink:  video.Title,
		Uploader:   video.Author,
		Timestamp:  video.PublishDate,
		Title:      video.Title,
		Thumbnails: nil,
		Duration:   video.Duration,
		Formats:    formats,
	}
	return &info, nil
}
