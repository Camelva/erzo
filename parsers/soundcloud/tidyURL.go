package soundcloud

import (
	"fmt"
	"github.com/camelva/erzo/parsers"
	"net/url"
	"regexp"
)

type urlType byte

const (
	typePlaylist urlType = iota + 1
	typeStation
	typeUser
	typeSong
)

func (t urlType) String() string {
	switch t {
	case typePlaylist:
		return "playlist"
	case typeStation:
		return "station"
	case typeUser:
		return "user"
	case typeSong:
		return "song"
	default:
		return "undefined"
	}
}

func (c *Client) tidyURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", parsers.ErrURLParse{URL: uri, Message: err.Error()}
	}

	// get real url after redirects
	if u.Host == "soundcloud.app.goo.gl" {
		resp, err := c.Get(u.String(), false)
		if err != nil {
			return "", err
		}
		resp.Body.Close()
		u = resp.Request.URL
	}

	// rebuildURL update existing url.URL object
	if err := rebuildURL(u); err != nil {
		return "", err
	}

	u.RawQuery = ""

	return u.String(), nil
}

func rebuildURL(u *url.URL) error {
	kind, data := parseURL(u)
	if kind != typeSong {
		return parsers.ErrFormatNotSupported(kind.String())
	}

	songParts := struct {
		user, title, secret string
	}{}
	var finalURL string

	switch len(data) {
	case 4:
		songParts.secret = data[3]
		fallthrough
	case 3:
		songParts.title = data[2]
		fallthrough
	case 2:
		songParts.user = data[1]
	}

	switch kind {
	case typeStation:
		finalURL = fmt.Sprintf("https://soundcloud.com/stations/track/%s/%s", songParts.user, songParts.title)
	case typePlaylist:
		finalURL = fmt.Sprintf("https://soundcloud.com/sets/%s/%s", songParts.user, songParts.title)
	case typeUser:
		finalURL = fmt.Sprintf("https://soundcloud.com/%s", songParts.user)
	case typeSong:
		finalURL = fmt.Sprintf("https://soundcloud.com/%s/%s", songParts.user, songParts.title)
	}

	if finalURL == "" {
		// impossible, but lets check anyways
		return parsers.ErrFormatNotSupported(kind.String())
	}

	if songParts.secret != "" {
		finalURL = fmt.Sprintf("%s/%s", finalURL, songParts.secret)
	}

	newURL, err := url.Parse(finalURL)
	if err != nil {
		return parsers.ErrURLParse{URL: finalURL, Message: err.Error()}
	}
	u = newURL
	return nil
}

func parseURL(u *url.URL) (kind urlType, result []string) {
	urlPath := u.EscapedPath()

	// Different SoundCloud url types
	stationRE := regexp.MustCompile(`^/(?:stations)/(?:track)/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`)
	playlistRE := regexp.MustCompile(`^/([\w-]+)/(?:sets)/([\w-]+)(?:|/|/([\w-]+)/?)$`)
	userRE := regexp.MustCompile(`^/([\w-]+)/?$`)
	songRE := regexp.MustCompile(`^/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`)

	urlTypes := map[urlType]*regexp.Regexp{
		typeStation:  stationRE,
		typePlaylist: playlistRE,
		typeUser:     userRE,
		typeSong:     songRE,
	}

	for name, pattern := range urlTypes {
		result := pattern.FindStringSubmatch(urlPath)
		if result == nil {
			continue
		}
		return name, result
	}
	return
}
