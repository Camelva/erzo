package loaders

import (
	"erzo/parsers"
	"net/url"
)

type Loader interface {
	Name() string
	Bin() string
	Get(*url.URL, string) error
	Compatible(format parsers.Format) bool
}
