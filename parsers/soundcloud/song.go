package soundcloud

import (
	"fmt"
	"github.com/camelva/erzo/parsers"
	"regexp"
	"time"
)

type Song struct {
	ID          string
	Streams     []Stream
	Title       string
	Author      string
	Duration    time.Duration
	PublishDate time.Time
	Thumbnails  map[string]parsers.Artwork
}

type Stream struct {
	Preset,
	URL,
	MimeType,
	Quality string
}

func (s *Song) FindStreamByQuality(quality string) *Stream {
	for i := range s.Streams {
		if s.Streams[i].Quality == quality {
			return &s.Streams[i]
		}
	}

	return nil
}

func (s *Song) FindStreamByPreset(preset string) *Stream {
	for i := range s.Streams {
		if s.Streams[i].Preset == preset {
			return &s.Streams[i]
		}
	}
	return nil
}

func (s *Song) parseSongInfo(meta *metadata2) error {
	s.ID = meta.Permalink
	s.Title = meta.Title
	s.Author = meta.User.Username

	duration, _ := time.ParseDuration(fmt.Sprintf("%dms", meta.Duration))
	s.Duration = duration.Round(time.Second)

	s.PublishDate = time.Date(meta.CreatedAt.Year(), meta.CreatedAt.Month(), meta.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

func (s *Song) addArtworks(meta *metadata2) {
	artwork := meta.ArtworkURL
	artworksMap := map[string]int{
		"mini":     16,
		"tiny":     20,
		"small":    32,
		"badge":    47,
		"t67x67":   67,
		"large":    100,
		"t300x300": 300,
		"crop":     400,
		"t500x500": 500,
		"original": 0,
	}
	if len(artwork) < 1 {
		artwork = meta.User.AvatarURL
	}

	artworks := make(map[string]parsers.Artwork, 0)

	re := regexp.MustCompile(`-([0-9a-z]+)\.jpg`)
	if !re.MatchString(artwork) {
		// no artworks, return empty slice
		return
	}

	for artType, artSize := range artworksMap {
		newType := fmt.Sprintf("-%s.jpg", artType)
		newURL := re.ReplaceAllString(artwork, newType)
		var i = parsers.Artwork{
			Type: artType,
			URL:  newURL,
			Size: artSize,
		}
		artworks[artType] = i
	}

	s.Thumbnails = artworks
}
