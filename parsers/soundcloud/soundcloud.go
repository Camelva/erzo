package soundcloud

import (
	"encoding/json"
	"erzo/types"
	"erzo/utils"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
)

type SoundCloudIE struct {
	urlPattern string
	apiURL     string
	api2URL    string
	baseURL    string
	clientID   string
}

type kind int8

const (
	songType kind = iota
	playlistType
	stationType
	userType
)

type scURL struct {
	title  string
	user   string
	kind   kind
	secret string
	url    string
}

var extractor = SoundCloudIE{
	urlPattern: `(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`,
	apiURL:     "https://api.soundcloud.com/",
	api2URL:    "https://api-v2.soundcloud.com/",
	baseURL:    "https://soundcloud.com/",
	clientID:   "psT32GLDMZ0TQKgfPkzrGIlco3PYA1kf",
}

func Init() SoundCloudIE {
	return extractor
}

func (ie SoundCloudIE) Compatible(s string) bool {
	ok, err := regexp.MatchString(extractor.urlPattern, s)
	if err != nil {
		log.Printf("[soundcloud] error while comparing: %s", err)
		return false
	}
	return ok
}

func (ie SoundCloudIE) Extract(u url.URL) (*types.ExtractorInfo, error) {
	sc := parseURL(u)
	if sc.kind != songType {
		err := fmt.Errorf("[soundcloud] playlists not supported yet")
		log.Print(err)
		return nil, err
	}
	metadata, err := resolve(sc.url)
	if err != nil {
		return nil, err
	}
	info, err := extractInfo(metadata)
	if err != nil {
		log.Printf("[soundcloud] error while extracting info: %s", err)
		return nil, err
	}
	return info, nil
}

func parseURL(u url.URL) *scURL {
	path := u.EscapedPath()
	if sc := parseStation(path); sc != nil {
		return sc
	}
	if sc := parsePlaylist(path); sc != nil {
		return sc
	}
	if sc := parseSong(path); sc != nil {
		return sc
	}
	// TODO: parse User link
	_ = userType
	return nil
}

func parseStation(path string) *scURL {
	pattern := `^[\/](?:stations)[\/]` + // /stations/
		`(?:track)[\/]` + // track/
		`([\w-]+)[\/]` + // user/
		`([\w-]+)[\/]?$` // title/
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(path)
	if len(result) < 1 {
		return nil
	}
	var (
		user   = result[1]
		title  = result[2]
		secret = result[3]
		uri    = fmt.Sprintf("%sstations/track/%s/%s", extractor.baseURL, user, title)
	)
	if len(secret) > 0 {
		uri += fmt.Sprintf("/%s", secret)
	}
	sc := scURL{
		title:  title,
		user:   user,
		kind:   stationType,
		secret: secret,
		url:    uri,
	}
	return &sc
}

func parsePlaylist(path string) *scURL {
	pattern := `^[\/]([\w-]+)[\/]` +
		`(?:sets)` +
		`[\/]([\w-]+)[\/]?` +
		`([\w-]+)?[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(path)
	if len(result) < 1 {
		return nil
	}
	var (
		user   = result[1]
		title  = result[2]
		secret = result[3]
		uri    = fmt.Sprintf("%ssets/%s/%s", extractor.baseURL, user, title)
	)
	if len(secret) > 0 {
		uri += fmt.Sprintf("/%s", secret)
	}
	sc := scURL{
		title:  title,
		user:   user,
		kind:   playlistType,
		secret: secret,
		url:    uri,
	}
	return &sc
}

func parseSong(path string) *scURL {
	pattern := `^[\/]([\w-]+)[\/]` +
		`([\w-]+)[\/]?` +
		`([\w-]+)?[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(path)
	if len(result) < 1 {
		return nil
	}
	var (
		user   = result[1]
		title  = result[2]
		secret = result[3]
		uri    = fmt.Sprintf("%s%s/%s", extractor.baseURL, user, title)
	)
	if len(secret) > 0 {
		uri += fmt.Sprintf("/%s", secret)
	}
	sc := scURL{
		title:  title,
		user:   user,
		kind:   songType,
		secret: secret,
		url:    uri,
	}
	return &sc
}

func resolve(link string) (*types.SoundCloudMetadata2, error) {
	resolveURL, err := url.Parse(fmt.Sprintf("%sresolve?url=%s", extractor.api2URL, link))
	if err != nil {
		log.Printf("[soundcloud] error while building resolve link: %s", err)
		return nil, err
	}
	res, err := fetch(resolveURL)
	if err != nil {
		return nil, err
	}
	if len(res) < 1 {
		log.Printf("%s\n", res)
		// TODO: update token
		err := fmt.Errorf("[soundcloud] token outdated")
		log.Println(err)
		return nil, err
	}
	var scMetadata = new(types.SoundCloudMetadata2)
	if err := json.Unmarshal(res, &scMetadata); err != nil {
		log.Printf("[soundcloud] error while unmarshaling metadata response: %s", err)
		log.Printf("%s", res)
		return nil, err
	}
	return scMetadata, nil
}

func extractInfo(info *types.SoundCloudMetadata2) (*types.ExtractorInfo, error) {
	trackID := info.ID
	formats := make(types.Formats, 0)
	//baseURL := fmt.Sprintf("%stracks/%d", sc.Host, trackID)
	//query := url.Values{"client_id" : {sc.ClientID}}
	if info.Downloadable && info.HasDownloadsLeft {
		dlURL, err := url.Parse(info.DownloadURL)
		if err != nil {
			return nil, err
		}
		q := dlURL.Query()
		q.Set("client_id", extractor.clientID)
		query := q.Encode()
		dlURL.RawQuery = query
		format := map[string]string{
			"url":      dlURL.String(),
			"ext":      "mp3",
			"type":     "mpeg",
			"protocol": "http",
			"score":    "100",
		}
		formats = append(formats, format)
	}

	if len(formats) < 1 {
		var err error
		transcodings := info.Media.Transcodings
		formats, err = extractFormats(transcodings)
		if err != nil {
			return nil, err
		}
	}

	duration := float32(info.Duration) * 1 / 1000

	thumbnails, err := extractArtworks(info.ArtworkURL, info.User.AvatarURL)
	if err != nil {
		thumbnails = make([]types.Artwork, 0)
	}

	var ExtractedInfo = &types.ExtractorInfo{
		ID:           trackID,
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

func extractFormats(transcodings []types.SoundCloudTranscoding) (types.Formats, error) {
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

		addFormat(&formats, t)
	}
	sortFormats(&formats)
	return formats, nil
}

func addFormat(formats *types.Formats, t types.SoundCloudTranscoding) {
	re := regexp.MustCompile(`_`)
	ext := re.Split(t.Preset, -1)[0]
	re = regexp.MustCompile(`audio/([\w-]+)[;]?`)
	mimeType := re.FindStringSubmatch(t.Format.MimeType)[1]
	f := map[string]string{
		"url":      t.URL,
		"type":     mimeType,
		"protocol": t.Format.Protocol,
		"ext":      ext,
	}
	*formats = append(*formats, f)
}

func sortFormats(formats *types.Formats) {
	formatsCopy := make(types.Formats, len(*formats))
	copy(formatsCopy, *formats)
	for i, format := range formatsCopy {
		var score int
		switch format["ext"] {
		case "mp3":
			score += 10
		case "opus":
			score += 5
		default:
			score += 0
		}
		switch format["protocol"] {
		case "progressive":
			score += 10
		case "hls":
			score += 5
		default:
			score += 0
		}
		formatsCopy[i]["score"] = strconv.Itoa(score)
	}
	sort.Slice(formatsCopy, func(i, j int) bool { return formatsCopy[i]["score"] > formatsCopy[j]["score"] })
	*formats = formatsCopy
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
		return nil, fmt.Errorf("there is no artworks")
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
	q := u.Query()
	if cID := q.Get("client_id"); cID == "" {
		q.Set("client_id", extractor.clientID)
		u.RawQuery = q.Encode()
	}
	res, err := utils.Fetch(u)
	if err != nil {
		log.Printf("[soundcloud] error while fetching url: %s", err)
		return nil, err
	}
	return res, nil
}
