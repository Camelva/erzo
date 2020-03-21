package soundcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"

	"erzo/types"
	"erzo/utils"
)

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

const tokenFile = "parsers/soundcloud/token.txt"

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

	tokenBytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Printf("reading token file: %s\n", err)
		return
	}
	tokenStr := string(tokenBytes)
	if len(tokenStr) == 32 {
		IE.clientID = tokenStr
	}
	return
}

var IE extractor

func (ie extractor) Compatible(s string) bool {
	ok, err := regexp.MatchString(IE.urlPattern, s)
	if err != nil {
		log.Printf("[soundcloud] error while comparing: %s", err)
		return false
	}
	return ok
}

func (ie extractor) Extract(u url.URL) (*types.ExtractorInfo, error) {
	sc := parseURL(u)
	if sc.kind != _song {
		err := types.ErrNotSupported{Subject: sc.kind.String()}
		log.Printf("[soundcloud] %s\n", err)
		return nil, err
	}
	metadata, err := resolve(sc.url)
	if err != nil {
		return nil, err
	}
	info, err := extractInfo(metadata)
	if err != nil {
		log.Printf("[soundcloud] extracting info: %s\n", err)
		return nil, err
	}
	return info, nil
}

func parseURL(u url.URL) *scURL {
	path := u.EscapedPath()
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
		result := k.FindStringSubmatch(path)
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
			log.Println("station")
			uri = fmt.Sprintf("%sstations/track/%s/%s", IE.baseURL, user, title)
		case _playlist:
			log.Println("playlist")
			uri = fmt.Sprintf("%ssets/%s/%s", IE.baseURL, user, title)
		case _user:
			log.Println("user")
			uri = fmt.Sprintf("%s%s", IE.baseURL, user)
		case _song:
			if user == "stations" {
				continue
			}
			if title == "sets" {
				continue
			}
			log.Println("song")
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
	resolveURL, err := url.Parse(fmt.Sprintf("%sresolve?url=%s", IE.api2URL, link))
	if err != nil {
		log.Printf("[soundcloud] building resolve link: %s\n", err)
		return nil, err
	}
	res, err := fetch(resolveURL)
	if err != nil {
		return nil, err
	}
	var scMetadata = new(metadata2)
	if err := json.Unmarshal(res, &scMetadata); err != nil {
		log.Printf("[soundcloud] unmarshalling metadata response: %s\nResponse: %s\n", err, res)
		return nil, err
	}
	return scMetadata, nil
}

func extractInfo(info *metadata2) (*types.ExtractorInfo, error) {
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

	thumbnails, err := extractArtworks(info.ArtworkURL, info.User.AvatarURL)
	if err != nil {
		log.Printf("[soundcloud] extracting artworks: %s\n", err)
	}

	var ExtractedInfo = &types.ExtractorInfo{
		ID:           info.ID,
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

func (info *metadata2) getDownloadLink() (types.Formats, bool) {
	if !info.Downloadable || !info.HasDownloadsLeft {
		return nil, false
	}
	dlURL, err := url.Parse(info.DownloadURL)
	if err != nil {
		log.Printf("[soundcloud] parsing download url: %s\n", err)
		return nil, false
	}
	q := dlURL.Query()
	q.Set("client_id", IE.clientID)
	query := q.Encode()
	dlURL.RawQuery = query
	format := types.Format{
		Url:      dlURL.String(),
		Ext:      "mp3",
		Type:     "mpeg",
		Protocol: "http",
		Score:    100,
	}
	return []types.Format{format}, true
}

func (transcodings transcodings) extractFormats() (types.Formats, error) {
	formats := make(types.Formats, 0)
	for _, t := range transcodings {
		formatURL, err := url.Parse(t.URL)
		if err != nil {
			return nil, err
		}

		stream, err := fetch(formatURL)
		if err != nil {
			return nil, err
		}

		var streamObj struct {
			URL string `json:"url"`
		}
		if err = json.Unmarshal(stream, &streamObj); err != nil {
			return nil, err
		}

		t.URL = streamObj.URL

		formats.Add(t)
	}
	formats.Sort()
	return formats, nil
}

func extractArtworks(artwork string, avatar string) ([]types.Artwork, error) {
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

	artworks := make([]types.Artwork, 0)

	re := regexp.MustCompile(`-([0-9a-z]+)\.jpg`)
	if !re.MatchString(artwork) {
		return artworks, fmt.Errorf("there is no artworks")
	}

	for artType, artSize := range artworksMap {
		newType := fmt.Sprintf("-%s.jpg", artType)
		newURL := re.ReplaceAllString(artwork, newType)
		var i = types.Artwork{
			Type: artType,
			URL:  newURL,
			Size: artSize,
		}
		artworks = append(artworks, i)
	}

	return artworks, nil
}

func fetch(u *url.URL) ([]byte, error) {
	// loop for two tries
	for range []int{0, 0} {
		q := u.Query()
		q.Set("client_id", IE.clientID)
		u.RawQuery = q.Encode()
		res, err := utils.Fetch(u)
		if err != nil {
			log.Printf("[soundcloud] fetching url: %s\n", err)
			return nil, err
		}
		if len(res) < 1 {
			if err := updateToken(); err != nil {
				return nil, err
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
		log.Printf("[soundcloud] fetching homepage: %s\n", err)
		return err
	}
	scriptTmpl := `<script[^>]+src="([^"]+)"`
	clientTmpl := `client_id\s*:\s*"([0-9a-zA-Z]{32})"`
	scriptRE := regexp.MustCompile(scriptTmpl)
	clientRE := regexp.MustCompile(clientTmpl)
	scripts := scriptRE.FindAllSubmatch(res, -1)
	for _, script := range scripts {
		scriptURL, err := url.Parse(string(script[1]))
		if err != nil {
			log.Printf("[soundcloud] parsing script url: %s\n", err)
			continue
		}
		scriptBody, err := utils.Fetch(scriptURL)
		if err != nil {
			log.Printf("[soundcloud] fetching script: %s\n", err)
			continue
		}
		matches := clientRE.FindSubmatch(scriptBody)
		if matches == nil {
			continue
		}
		IE.clientID = string(matches[1])
		if err := ioutil.WriteFile(tokenFile, matches[1], 0644); err != nil {
			log.Printf("[soundcloud] updating token file: %s\n", err)
		}
		log.Println(IE.clientID)
		return nil
	}
	return fmt.Errorf("can't retrieve token")
}
