package main

import (
	"bufio"
	"encoding/json"
	"erzo/loader"
	"erzo/parsers"
	"fmt"
	"net/url"
	"os"
	"regexp"
)

const (
	urlPattern = `((?:[a-z]{3,6}:\/\/)|(?:^|\s))` +
		`((?:[a-zA-Z0-9\-]+\.)+[a-z]{2,13})` +
		`([\.\?\=\&\%\/\w\-]*\b)`
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter link: ")
	userInput, _ := reader.ReadString('\n')
	r, err := Get(userInput)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	//_ = r
	fmt.Printf("Response: %s", r)
}

func Get(message string) (string, error) {
	urlObj, err := extractURL(message)
	if err != nil {
		return "", err
	}

	info, err := parsers.Parse(urlObj)
	if err != nil {
		return "", err
	}

	//fmt.Printf("%+v", info)
	if err := PrettyPrint(info); err != nil {
		return "", err
	}

	for _, format := range info.Formats {
		fmt.Println(format["url"])
	}
	
	fileName, err := loader.Go(info.Formats)
	if err != nil {
		return "", err
	}

	return fileName, nil
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

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}