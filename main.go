package main

import (
	"erzo/parsers"
	"fmt"
	"net/url"
	"regexp"
)

const (
	urlPattern = `((?:[a-z]{3,6}:\/\/)|(?:^|\s))` +
		`((?:[a-zA-Z0-9\-]+\.)+[a-z]{2,13})` +
		`([\.\?\=\&\%\/\w\-]*\b)`
)

func main() {
	r, err := Get("some text with url " +
		"https://soundcloud.com/bonsaicollct/colson-xl-torrid " +
		"and more text")
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	fmt.Printf("Response: %s", r)
}

func Get(message string) (res string, err error) {
	urlObj, err := extractURL(message)
	if err != nil {
		return "", err
	}

	// TODO: improve this
	isSC, err := regexp.MatchString(`(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`, urlObj.Hostname())
	if err != nil {
		return "", err
	}
	if isSC {
		dlURL, err := soundcloud.Get(urlObj)
		if err != nil {
			return "", err
		}
		res = dlURL
	}

	return urlObj.String(), nil
}

func extractURL(message string) (*url.URL, error) {
	re := regexp.MustCompile(urlPattern)
	rawURL := re.FindString(message)
	link, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return link, nil
}
