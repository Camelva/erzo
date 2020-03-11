package parsers

import (
	"encoding/json"
	"erzo/types"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
)

const (
	SoundCloudPattern = `(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`
)

type SoundCloud struct {
	types.Soundcloud
}

func (sc *SoundCloud) Get(u *url.URL) (*types.ExtractorInfo, error) {
	err := sc.parse(u)
	if err != nil {
		return nil, err
	}
	info, err := sc.extract()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (sc *SoundCloud) parse(u *url.URL) error {
	sc.init()
	err := sc.parsePath(u.EscapedPath())
	if err != nil {
		return err
	}
	return nil
}

func (sc *SoundCloud) init() {
	const (
		clientID = "uY5YAJMT1mVRQtgKBoNbJqlVGILaYI0p"
		scAPI    = "https://api.soundcloud.com/"
		scAPIv2  = "https://api-v2.soundcloud.com/"
		scBase   = "https://soundcloud.com/"
	)

	sc.ClientID = clientID
	sc.Host = scBase
	sc.API = scAPI
	sc.APIv2 = scAPIv2
	return
}

func (sc *SoundCloud) parsePath(path string) error {
	if sc.parseStation(path) {
		return nil
	} else if sc.parsePlaylist(path) {
		return nil
	} else if sc.parseSong(path) {
		return nil
	}
	return fmt.Errorf("can't parse this link")
}

func (sc *SoundCloud) parseStation(s string) bool {
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

func (sc *SoundCloud) parsePlaylist(s string) bool {
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

func (sc *SoundCloud) parseSong(s string) bool {
	pattern := `^[\/]([\w-]+)[\/]([\w-]+)[\/]?([\w-]+)?[\/]?$`
	re := regexp.MustCompile(pattern)
	result := re.FindStringSubmatch(s)
	if len(result) < 1 {
		return false
	}
	sc.User = result[1]
	sc.Title = result[2]
	if result[3] != "" {
		sc.Secret = result[3]
	}
	return true
}

func (sc *SoundCloud) extract() (*types.ExtractorInfo, error) {
	metadata, err := sc.resolve()
	if err != nil {
		return nil, err
	}
	info, err := sc.extractInfo(metadata)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (sc *SoundCloud) resolve() (*types.SoundCloudMetadata2, error) {
	resolveURL, err := sc.buildResolveLink()
	if err != nil {
		return nil, err
	}
	b, err := sc.fetchURL(resolveURL)
	if err != nil {
		return nil, err
	}
	var scMetadata = new(types.SoundCloudMetadata2)
	if err := json.Unmarshal(b, &scMetadata); err != nil {
		return nil, err
	}
	return scMetadata, nil
}

func (sc *SoundCloud) buildResolveLink() (*url.URL, error) {
	if sc.URL == "" {
		sc.buildLink()
	}
	u := fmt.Sprintf("%sresolve?url=%s", sc.APIv2, sc.URL)
	urlObj, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return urlObj, nil
}

func (sc *SoundCloud) buildLink() {
	if sc.Playlist {
		sc.URL = fmt.Sprintf("%s%s/sets/%s", sc.Host, sc.User, sc.Title)
		return
	} else if sc.Station {
		sc.URL = fmt.Sprintf("%sstations/track/%s/%s", sc.Host, sc.User, sc.Title)
		return
	}
	sc.URL = fmt.Sprintf("%s%s/%s", sc.Host, sc.User, sc.Title)

	if sc.Secret == "" {
		return
	}
	secretPath := fmt.Sprintf("/%s", sc.Secret)
	sc.URL += secretPath
	return
}

func (sc *SoundCloud) extractInfo(info *types.SoundCloudMetadata2) (*types.ExtractorInfo, error) {
	trackID := info.ID
	formats := make([]map[string]string, 0)
	//baseURL := fmt.Sprintf("%stracks/%d", sc.Host, trackID)
	//query := url.Values{"client_id" : {sc.ClientID}}
	if info.Downloadable && info.HasDownloadsLeft {
		dlURL, err := url.Parse(info.DownloadURL)
		if err != nil { return nil, err }
		q := dlURL.Query()
		q.Set("client_id", sc.ClientID)
		query := q.Encode()
		dlURL.RawQuery = query
		format := map[string]string{
			"url": dlURL.String(),
			"ext": "mp3",
			"type": "mpeg",
			"protocol": "http",
			"score": "100",
		}
		formats = append(formats, format)
	}

	if len(formats) < 1 {
		var err error
		transcodings := info.Media.Transcodings
		formats, err = sc.extractFormats(transcodings)
		if err != nil { return nil, err }
	}

	duration := float32(info.Duration) * 1 / 1000

	thumbnails, err := sc.extractArtwork(info.ArtworkURL, info.User.AvatarURL)
	if err != nil { return nil, err }

	var ExtractedInfo = &types.ExtractorInfo{
		ID: trackID,
		Uploader: info.User.Username,
		UploaderID: info.User.ID,
		UploaderURL: info.User.PermalinkURL,
		Timestamp: info.CreatedAt,
		Title: info.Title,
		Description: info.Description,
		Thumbnails: thumbnails,
		Duration: duration,
		WebPageURL: info.PermalinkURL,
		License: info.License,
		ViewCount: info.PlaybackCount,
		LikeCount: info.LikesCount,
		CommentCount: info.CommentCount,
		RepostCount: info.RepostsCount,
		Genre: info.Genre,
		Formats: formats,
	}

	return ExtractedInfo, nil
}

func (sc *SoundCloud) extractFormats(transcodings []types.SoundCloudTranscoding) ([]map[string]string, error) {
	formats := make([]map[string]string, 0)
	for _, t := range transcodings {
		formatURL, err := url.Parse(t.URL)
		if err != nil { return nil, err }

		stream, err := sc.fetchURL(formatURL)
		if err != nil { return nil, err }

		var streamObj struct { URL string `json:"url"` }
		if err = json.Unmarshal(stream, &streamObj); err != nil {
			return nil, err
		}

		t.URL = streamObj.URL

		sc.addFormat(&formats, t)
	}
	sc.sortFormats(&formats)
	return formats, nil
}

func (sc *SoundCloud) addFormat(formats *[]map[string]string, t types.SoundCloudTranscoding) {
	re := regexp.MustCompile(`_`)
	ext := re.Split(t.Preset, -1)[0]
	re = regexp.MustCompile(`audio/([\w-]+)[;]?`)
	mimeType := re.FindStringSubmatch(t.Format.MimeType)[1]
	f := map[string]string{
		"url": t.URL,
		"type": mimeType,
		"protocol": t.Format.Protocol,
		"ext": ext,
	}
	*formats = append(*formats, f)
}

func (sc *SoundCloud) sortFormats(formats *[]map[string]string) {
	formatsCopy := make([]map[string]string, 3)
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

func (sc *SoundCloud) extractArtwork(artwork string, avatar string) ([]types.Artwork, error) {
	artworksMap := map[string]int{
		"mini": 16,
		"tiny": 20,
		"small": 32,
		"badge": 47,
		"t67x67": 67,
		"large": 100,
		"t300x300": 300,
		"crop": 400,
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

func (sc *SoundCloud) fetchURL(u *url.URL) ([]byte, error) {
	q := u.Query()
	if cID := q.Get("client_id"); cID == "" {
		q.Set("client_id", sc.ClientID)
		u.RawQuery = q.Encode()
	}
	res, err := fetchURL(u)
	if err != nil {
		return nil, err
	}
	return res, nil
}
