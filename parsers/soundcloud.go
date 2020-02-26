package soundcloud

import (
	"fmt"
	"net/url"
	"regexp"
)

const (
	clientID = "uY5YAJMT1mVRQtgKBoNbJqlVGILaYI0p"
	scAPI    = "https://api.soundcloud.com/"
	scAPIv2  = "https://api-v2.soundcloud.com/"
	scBase   = "https://soundcloud.com/"
)

type Soundcloud struct {
	Host     string
	API      string
	APIv2    string
	URL      string
	User     string
	Title    string
	Playlist bool
	Station  bool
	ClientID string
}

func Get(u *url.URL) (string, error) {
	sc, err := parse(u)
	if err != nil {
		return "", err
	}
	rl := sc.buildResolveLink()
	fmt.Println(rl)
	fmt.Printf("%+v\n", sc)
	return "uri", nil
}

func parse(u *url.URL) (*Soundcloud, error) {
	sc := new(Soundcloud)
	sc.init()
	err := sc.parsePath(u.EscapedPath())
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func (sc *Soundcloud) init() {
	sc.ClientID = clientID
	sc.Host = scBase
	sc.API = scAPI
	sc.APIv2 = scAPIv2
	return
}

func (sc *Soundcloud) parsePath(path string) error {
	if sc.parseStation(path) {
		return nil
	} else if sc.parsePlaylist(path) {
		return nil
	} else if sc.parseSong(path) {
		return nil
	}
	return fmt.Errorf("can't parse this link")
}

func (sc *Soundcloud) parseStation(s string) bool {
	pattern := `^[\/](?:stations)[\/](?:track)[\/]([\w-]+)\/([\w-]+)[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(s)
	if len(result) < 1 {
		return false
	}
	sc.Station = true
	sc.User = result[1]
	sc.Title = result[2]
	return true
}

func (sc *Soundcloud) parsePlaylist(s string) bool {
	pattern := `^[\/]([\w-]+)[\/](?:sets)[\/]([\w-]+)[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(s)
	if len(result) < 1 {
		return false
	}
	sc.Playlist = true
	sc.User = result[1]
	sc.Title = result[2]
	return true
}

func (sc *Soundcloud) parseSong(s string) bool {
	pattern := `^[\/]([\w-]+)[\/]([\w-]+)[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(s)
	if len(result) < 1 {
		return false
	}
	sc.User = result[1]
	sc.Title = result[2]
	return true
}

func (sc *Soundcloud) buildLink() {
	if sc.Playlist {
		sc.URL = fmt.Sprintf("%s%s/sets/%s", sc.Host, sc.User, sc.Title)
		return
	} else if sc.Station {
		sc.URL = fmt.Sprintf("%sstations/track/%s/%s", sc.Host, sc.User, sc.Title)
		return
	}
	sc.URL = fmt.Sprintf("%s%s/%s", sc.Host, sc.User, sc.Title)
	return
}

func (sc *Soundcloud) buildResolveLink() string {
	if sc.URL == "" {
		sc.buildLink()
	}
	return fmt.Sprintf("%sresolve?url=%s&client_id=%s", sc.APIv2, sc.URL, sc.ClientID)
}
