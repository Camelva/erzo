package soundcloud

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
		urlPattern: `(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`,
		name:       "SoundCloud",
	}
	engine.AddExtractor(IE)
}

func (ie Extractor) Name() string {
	return ie.name
}

func (ie Extractor) Compatible(u *url.URL) bool {
	s := u.Hostname()
	if s == "soundcloud.app.goo.gl" {
		return true
	}
	ok, _ := regexp.MatchString(ie.urlPattern, s)
	return ok
}

func (ie Extractor) Extract(u *url.URL, debug bool, client *http.Client) (*parsers.ExtractorInfo, error) {
	c := &Client{Debug: debug, HTTPClient: client}
	song, err := c.GetSong(u.String())
	if err != nil {
		return nil, err
	}

	formats := parsers.Formats{}

	if len(song.Streams) < 1 {
		return nil, parsers.ErrCantContinue("found no streams")
	}

	for _, stream := range song.Streams {
		f := parsers.Format{
			Url:      stream.URL,
			Ext:      "",
			Type:     stream.MimeType,
			Protocol: "https",
			Score:    0,
		}
		formats = append(formats, f)
	}

	info := parsers.ExtractorInfo{
		Permalink:  song.ID,
		Uploader:   song.Author,
		Timestamp:  song.PublishDate,
		Title:      song.Title,
		Thumbnails: song.Thumbnails,
		Duration:   song.Duration,
		Formats:    formats,
	}
	return &info, nil
}
