package soundcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/camelva/erzo/parsers"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Client offers methods to download video metadata and video streams.
type Client struct {
	// Debug enables debugging output through log package
	Debug bool

	// HTTPClient can be used to set a custom HTTP client.
	// If not set, http.DefaultClient will be used
	HTTPClient *http.Client

	// ClientID used for making requests to SoundCloud
	ClientID string
}

// GetVideo fetches video metadata
func (c *Client) GetSong(uri string) (*Song, error) {
	return c.GetSongContext(context.Background(), uri)
}

// GetVideoContext fetches video metadata with a context
func (c *Client) GetSongContext(ctx context.Context, uri string) (*Song, error) {
	meta, err := c.getMetadata(ctx, uri)
	if err != nil {
		return nil, err
	}

	s := new(Song)
	if err := s.parseSongInfo(meta); err != nil {
		return nil, err
	}

	streams, err := c.parseStreams(meta)
	if err != nil {
		return nil, err
	}
	s.Streams = streams
	if len(s.Streams) == 0 {
		return nil, errors.New("no Stream list found in the server's answer")
	}
	return s, nil
}

func (c *Client) parseStreams(meta *metadata2) ([]Stream, error) {
	size := len(meta.Media.Transcodings)
	streams := make([]Stream, 0, size)
	if meta.Downloadable && meta.HasDownloadsLeft {
		stream, ok := c.getOriginalStream(meta.DownloadURL)
		if ok == true {
			streams = append(streams, *stream)
		}
	}

	transcodings := meta.Media.Transcodings
	extract := func(t transcoding) *Stream {
		streamResp, err := c.Get(t.URL, true)
		if err != nil {
			return nil
		}

		defer streamResp.Body.Close()
		stream, err := ioutil.ReadAll(streamResp.Body)
		if err != nil {
			return nil
		}

		var streamObj struct {
			URL string `json:"url"`
		}
		if err = json.Unmarshal(stream, &streamObj); err != nil {
			return nil
		}

		return &Stream{
			Preset:   t.Preset,
			URL:      streamObj.URL,
			MimeType: t.Format.MimeType,
			Quality:  t.Quality,
		}
	}

	for _, t := range transcodings {
		stream := extract(t)
		if stream != nil {
			streams = append(streams, *stream)
		}
	}
	return streams, nil
}

func (c *Client) getOriginalStream(uri string) (stream *Stream, ok bool) {
	resp, err := c.Get(uri, true)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false
	}
	var realDlURL struct {
		URL string `json:"redirectUri"`
	}
	if err = json.Unmarshal(respData, &realDlURL); err != nil {
		// invalid json, return false
		return nil, false
	}

	return &Stream{
		Preset:   "original",
		URL:      realDlURL.URL,
		MimeType: "mpeg",
		Quality:  "best",
	}, true
}

// GetStream returns the HTTP response for a specific stream
//func (c *Client) GetStream(song *Song, stream *Stream) (*http.Response, error) {
//	return c.GetStreamContext(context.Background(), song, stream)
//}

// GetStreamContext returns the HTTP response for a specific stream with a context
//func (c *Client) GetStreamContext(ctx context.Context, song *Song, stream *Stream) (*http.Response, error) {
//	uri, err := c.getStreamURL(ctx, song, stream)
//	if err != nil {
//		return nil, err
//	}
//
//	return c.httpGet(ctx, uri, true)
//}

//func (c *Client) GetStreamURL(song *Song, stream *Stream) (string, error) {
//	return c.getStreamURL(context.Background(), song, stream)
//}

//func (c *Client) getStreamURL(ctx context.Context, song *Song, stream *Stream) (string, error) {
//	if stream.URL != "" {
//		return stream.URL, nil
//	}
//
//	cipher := stream.Cipher
//	if cipher == "" {
//		return "", ErrCipherNotFound
//	}
//
//	return c.decipherURL(ctx, video.ID, cipher)
//}

func (c *Client) Get(url string, withClientID bool) (resp *http.Response, err error) {
	return c.httpGet(context.Background(), url, withClientID)
}

func (c *Client) httpGet(ctx context.Context, uri string, withClientID bool) (resp *http.Response, err error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	if c.Debug {
		log.Println("GET", uri)
	}

	if withClientID {
		if err := c.addClientID(&uri); err != nil {
			return nil, ErrNoClientID(err.Error())
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36")
	//req.Header.Set("Accept", "*/*")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	// Token expired
	if resp.StatusCode == http.StatusUnauthorized {
		if ctx.Value("repeated") == true {
			return nil, fmt.Errorf("can't make request to soundcloud.com")
		}
		if err := c.updateToken(); err != nil {
			return nil, fmt.Errorf("can't update token: %s", err.Error())
		}
		ctx = context.WithValue(ctx, "repeated", true)
		return c.httpGet(ctx, uri, true)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, ErrUnexpectedStatusCode(resp.StatusCode)
	}

	return
}

func (c *Client) addClientID(uri *string) error {
	u, err := url.Parse(*uri)
	if err != nil {
		return parsers.ErrURLParse{Message: err.Error(), URL: *uri}
	}

	q := u.Query()

	if c.ClientID == "" {
		t, err := readTokenFromFile()
		if err == nil {
			c.ClientID = t
			return nil
		}
		if err := c.updateToken(); err != nil {
			return err
		}
	}
	q.Set("client_id", c.ClientID)
	u.RawQuery = q.Encode()
	*uri = u.String()
	return nil
}
