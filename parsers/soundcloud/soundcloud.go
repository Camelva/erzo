package soundcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
	"github.com/camelva/erzo/utils"
)

var debugInstance = "soundcloud"

// type for identifying url kind
type urlKind int8

// const for identifying user-provided url's type (kind)
const (
	_song urlKind = iota
	_playlist
	_station
	_user
)

func (k urlKind) String() string {
	switch k {
	case 0:
		return "song"
	case 1:
		return "playlist"
	case 2:
		return "station"
	case 3:
		return "user"
	default:
		return "undefined"
	}
}

var tokenFile = path.Join(os.TempDir(), "soundcloud-token.txt")

var IE extractor

func init() {
	//noinspection SpellCheckingInspection
	clientIDBase := "psT32GLDMZ0TQKgfPkzrGIlco3PYA1kf"

	IE = extractor{
		urlPattern: `(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`,
		apiURL:     "https://api.soundcloud.com/",
		api2URL:    "https://api-v2.soundcloud.com/",
		baseURL:    "https://soundcloud.com/",
		clientID:   clientIDBase,
	}

	if tokenBytes, err := ioutil.ReadFile(tokenFile); err == nil {
		tokenStr := string(tokenBytes)
		if len(tokenStr) == 32 {
			IE.clientID = tokenStr
		}
	}
	engine.AddExtractor(IE)
	return
}

// Main struct with necessary info and methods
type extractor struct {
	urlPattern string
	apiURL     string
	api2URL    string
	baseURL    string
	clientID   string
}

func (ie extractor) Compatible(u url.URL) bool {
	s := u.Hostname()
	ok, _ := regexp.MatchString(ie.urlPattern, s)
	return ok
}

func (ie extractor) Extract(u url.URL) (*parsers.ExtractorInfo, error) {
	sc := parseURL(u)
	if sc.kind != _song {
		err := parsers.ErrNotSupported{Subject: sc.kind.String()}
		engine.Log(debugInstance, fmt.Errorf("extracting song info: %s", err))
		return nil, err
	}
	metadata, err := resolve(sc.url)
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		err := fmt.Errorf("can't get info")
		engine.Log(debugInstance, fmt.Errorf("resolving metadata: %s", err))
		return nil, err
	}
	info, err := extractInfo(metadata)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("extracting info: %s", err))
		return nil, err
	}
	return info, nil
}

// Struct containing info about user-provided url
type scURL struct {
	title  string
	user   string
	kind   urlKind
	secret string
	url    string
}

func parseURL(u url.URL) *scURL {
	urlPath := u.EscapedPath()
	stationTmpl := `^/(?:stations)/(?:track)/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	stationRE := regexp.MustCompile(stationTmpl)
	playlistTmpl := `^/([\w-]+)/(?:sets)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	playlistRE := regexp.MustCompile(playlistTmpl)
	userTmpl := `^/([\w-]+)/?$`
	userRE := regexp.MustCompile(userTmpl)
	songTmpl := `^/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	songRE := regexp.MustCompile(songTmpl)
	kinds := []*regexp.Regexp{_station: stationRE, _playlist: playlistRE, _user: userRE, _song: songRE}
	for t, k := range kinds {
		result := k.FindStringSubmatch(urlPath)
		if result == nil {
			continue
		}
		var user, title, secret, uri string
		if len(result) > 1 {
			user = result[1]
		}
		if len(result) > 2 {
			title = result[2]
		}
		if len(result) > 3 {
			secret = result[3]
		}

		switch urlKind(t) {
		case _station:
			uri = fmt.Sprintf("%sstations/track/%s/%s", IE.baseURL, user, title)
		case _playlist:
			uri = fmt.Sprintf("%ssets/%s/%s", IE.baseURL, user, title)
		case _user:
			uri = fmt.Sprintf("%s%s", IE.baseURL, user)
		case _song:
			if user == "stations" {
				continue
			}
			if title == "sets" {
				continue
			}
			uri = fmt.Sprintf("%s%s/%s", IE.baseURL, user, title)
		}
		sc := scURL{
			title:  title,
			user:   user,
			kind:   urlKind(t),
			secret: secret,
			url:    uri,
		}
		return &sc
	}
	return &scURL{}
}

func resolve(link string) (*metadata2, error) {
	_url := fmt.Sprintf("%sresolve?url=%s", IE.api2URL, link)
	resolveURL, err := url.Parse(_url)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("parsing url `%s`: %s", _url, err))
		return nil, err
	}
	res, err := fetch(resolveURL)
	if err != nil {
		return nil, err
	}
	var scMetadata = new(metadata2)
	if err := json.Unmarshal(res, &scMetadata); err != nil {
		engine.Log(debugInstance, fmt.Errorf("unmarshalling metadata: %s", err))
		return nil, err
	}
	return scMetadata, nil
}

func extractInfo(info *metadata2) (*parsers.ExtractorInfo, error) {
	formats, ok := info.getDownloadLink()
	if !ok {
		var err error
		transcodings := info.Media.Transcodings
		formats, err = transcodings.extractFormats()
		if err != nil {
			return nil, err
		}
	}

	duration := float32(info.Duration) * 1 / 1000

	thumbnails := extractArtworks(info.ArtworkURL, info.User.AvatarURL)

	var ExtractedInfo = &parsers.ExtractorInfo{
		ID:           info.ID,
		Permalink:    info.Permalink,
		Uploader:     info.User.Username,
		UploaderID:   info.User.ID,
		UploaderURL:  info.User.PermalinkURL,
		Timestamp:    info.CreatedAt,
		Title:        info.Title,
		Description:  info.Description,
		Thumbnails:   thumbnails,
		Duration:     duration,
		WebPageURL:   info.PermalinkURL,
		License:      info.License,
		ViewCount:    info.PlaybackCount,
		LikeCount:    info.LikesCount,
		CommentCount: info.CommentCount,
		RepostCount:  info.RepostsCount,
		Genre:        info.Genre,
		Formats:      formats,
	}

	return ExtractedInfo, nil
}

func (info *metadata2) getDownloadLink() (parsers.Formats, bool) {
	if !info.Downloadable || !info.HasDownloadsLeft {
		return nil, false
	}
	dlURL, err := url.Parse(info.DownloadURL)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("can't parse url `%s`. Error: %s", info.DownloadURL, err))
		return nil, false
	}
	q := dlURL.Query()
	q.Set("client_id", IE.clientID)
	query := q.Encode()
	dlURL.RawQuery = query
	format := parsers.Format{
		Url:      dlURL.String(),
		Ext:      "mp3",
		Type:     "mpeg",
		Protocol: "http",
		Score:    100,
	}
	return []parsers.Format{format}, true
}

func (transcodings transcodings) extractFormats() (parsers.Formats, error) {
	formats := make(parsers.Formats, 0)
	for _, t := range transcodings {
		formatURL, err := url.Parse(t.URL)
		if err != nil {
			engine.Log(debugInstance, fmt.Errorf("can't parse url `%s`. Error: %s", t.URL, err))
			continue
		}

		stream, err := fetch(formatURL)
		if err != nil {
			continue
		}

		var streamObj struct {
			URL string `json:"url"`
		}
		if err = json.Unmarshal(stream, &streamObj); err != nil {
			engine.Log(debugInstance, fmt.Errorf("unmarshalling streamObj: %s", err))
			continue
		}

		t.URL = streamObj.URL

		formats.Add(t)
	}
	formats.Sort()
	return formats, nil
}

func extractArtworks(artwork string, avatar string) []parsers.Artwork {
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
		artwork = avatar
	}

	artworks := make([]parsers.Artwork, 0)

	re := regexp.MustCompile(`-([0-9a-z]+)\.jpg`)
	if !re.MatchString(artwork) {
		err := fmt.Errorf("there is no artworks")
		engine.Log(debugInstance, fmt.Errorf("extracting artworks: %s", err))
		return artworks
	}

	for artType, artSize := range artworksMap {
		newType := fmt.Sprintf("-%s.jpg", artType)
		newURL := re.ReplaceAllString(artwork, newType)
		var i = parsers.Artwork{
			Type: artType,
			URL:  newURL,
			Size: artSize,
		}
		artworks = append(artworks, i)
	}

	return artworks
}

func fetch(u *url.URL) ([]byte, error) {
	// loop for two tries
	for i := range []int{0, 0} {
		q := u.Query()
		q.Set("client_id", IE.clientID)
		u.RawQuery = q.Encode()
		res, err := utils.Fetch(u)
		if err != nil {
			engine.Log(debugInstance, fmt.Errorf("fetching url `%s`. Error: %s", u.String(), err))
			continue
		}
		if len(res) < 1 && i == 0 {
			engine.Log(debugInstance, fmt.Errorf("updating token"))
			if err := updateToken(); err != nil {
				engine.Log(debugInstance, fmt.Errorf("while updationg token: %s", err))
				continue
			}
			continue
		}
		return res, nil
	}
	return nil, fmt.Errorf("can't fetch url")
}

func updateToken() error {
	u, _ := url.Parse("https://soundcloud.com")
	res, err := utils.Fetch(u)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("updateToken() can't fetch homepage: %s\n", err))
		return err
	}
	scriptTmpl := `<script[^>]+src="([^"]+)"`
	clientTmpl := `client_id\s*:\s*"([0-9a-zA-Z]{32})"`
	scriptRE := regexp.MustCompile(scriptTmpl)
	clientRE := regexp.MustCompile(clientTmpl)
	scripts := scriptRE.FindAllSubmatch(res, -1)
	for _, script := range scripts {
		uri := string(script[1])
		scriptURL, err := url.Parse(uri)
		if err != nil {
			engine.Log(debugInstance,
				fmt.Errorf("updateToken() can't parse script url `%s`. Error: %s", uri, err))
			continue
		}
		scriptBody, err := utils.Fetch(scriptURL)
		if err != nil {
			engine.Log(debugInstance,
				fmt.Errorf("updateToken() can't fetch script `%s`. Error: %s", scriptURL.String(), err))
			continue
		}
		matches := clientRE.FindSubmatch(scriptBody)
		if matches == nil {
			continue
		}
		IE.clientID = string(matches[1])
		if err := ioutil.WriteFile(tokenFile, matches[1], 0644); err != nil {
			engine.Log(debugInstance,
				fmt.Errorf("can't write new token to file `%s`: %s", tokenFile, err))
		}
		return nil
	}
	return fmt.Errorf("can't retrieve token")
}
