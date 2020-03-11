package parsers

import (
	"erzo/types"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func Parse(u *url.URL) (*types.ExtractorInfo, error) {
	if ok, _ := regexp.MatchString(SoundCloudPattern, u.Hostname()); ok {
		sc := new(SoundCloud)
		res, err := sc.Get(u)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}

func fetchURL(u *url.URL) ([]byte, error) {
	var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0"

	req, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}