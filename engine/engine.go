package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/camelva/erzo/loaders"
	"github.com/camelva/erzo/parsers"
)

var _extractors []parsers.Extractor
var _loaders []loaders.Loader

var _debug bool
var debugInstance = "engine"

const (
	_urlPattern = `((?:[a-z]{3,6}:\/\/)|(?:^|\s))` +
		`((?:[a-zA-Z0-9\-]+\.)+[a-z]{2,13})` +
		`([\.\?\=\&\%\/\w\-]*\b)`
)

var ErrNotURL = errors.New("there is no valid url")

func AddExtractor(x parsers.Extractor) {
	_extractors = append(_extractors, x)
}
func AddLoader(l loaders.Loader) {
	_loaders = append(_loaders, l)
}

type Engine struct {
	extractors   []parsers.Extractor
	loaders      []loaders.Loader
	outputFolder string
}

func New(out string, truncate bool, debug bool) *Engine {
	_debug = debug
	e := &Engine{
		extractors:   _extractors,
		loaders:      _loaders,
		outputFolder: out,
	}
	if truncate {
		e.clean()
	}
	return e
}

func (e Engine) Process(s string) (string, error) {
	u, ok := extractURL(s)
	if !ok {
		return "", ErrNotURL
	}
	info, err := e.extractInfo(*u)
	if err != nil {
		return "", err
	}
	title, err := e.downloadSong(info)
	if err != nil {
		return "", err
	}
	return title, nil
}

func (e Engine) extractInfo(u url.URL) (*parsers.ExtractorInfo, error) {
	for _, xtr := range e.extractors {
		if !xtr.Compatible(u) {
			continue
		}
		info, err := xtr.Extract(u)
		if err != nil {
			Log(debugInstance, fmt.Errorf("can't extract info: %s", err))
			return nil, err
		}
		return info, nil
	}
	return nil, parsers.ErrNotSupported{Subject: u.Hostname()}
}

func (e Engine) downloadSong(info *parsers.ExtractorInfo) (string, error) {
	if _, err := ioutil.ReadDir(e.outputFolder); err != nil {
		if err := os.Mkdir(e.outputFolder, 0644); err != nil {
			Log(debugInstance, fmt.Errorf("can't create folder"))
			e.outputFolder = ""
		}
	}
	outFile := fmt.Sprintf("%s.mp3", info.Permalink)
	outPath := path.Join(e.outputFolder, outFile)
	for _, format := range info.Formats {
		u, err := url.Parse(format.Url)
		if err != nil {
			Log(debugInstance, fmt.Errorf("can't parse format url: %s", err))
			continue
		}
		for _, ldr := range e.loaders {
			if !ldr.Compatible(format) {
				continue
			}
			if err := ldr.Get(u, outPath); err != nil {
				Log(debugInstance, fmt.Errorf("loader cant retrieve url: %s", err))
				continue
			}
			return outPath, nil
		}
	}
	return "", fmt.Errorf("unsupported protocol")
}

func (e Engine) clean() {
	if err := os.RemoveAll(e.outputFolder); err != nil {
		Log(debugInstance, fmt.Errorf("clean(): %s", err))
	}
	return
}

func extractURL(message string) (u *url.URL, ok bool) {
	re := regexp.MustCompile(_urlPattern)
	rawURL := re.FindString(message)
	if len(rawURL) < 1 {
		return nil, false
	}
	link, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("[engine] extractURL(): %s\n", err)
		return nil, false
	}
	return link, true
}

func Log(instance string, err error) {
	if !_debug {
		return
	}
	msg := fmt.Sprintf("[%s] %s", instance, err.Error())
	log.Println(msg)
}
