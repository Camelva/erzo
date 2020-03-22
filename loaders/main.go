package loaders

import (
	"net/url"

	"github.com/camelva/erzo/parsers"
)

type Loader interface {
	Name() string
	Bin() string
	Get(*url.URL, string) error
	Compatible(format parsers.Format) bool
}
