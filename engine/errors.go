package engine

import (
	"fmt"
	"github.com/camelva/erzo/parsers"
)

//var ErrNotURL = "there is no valid url"

type ErrNotURL struct{}

func (ErrNotURL) Error() string {
	return "there is no valid url"
}

type ErrUndefined struct{}

func (ErrUndefined) Error() string {
	return "undefined error"
}

// parsers errors
type ErrUnsupportedService struct {
	Service string
}

func (e ErrUnsupportedService) Error() string {
	return fmt.Sprintf("%s unsupported yet", e.Service)
}

type ErrUnsupportedType struct {
	parsers.ErrFormatNotSupported
}

type ErrCantFetchInfo struct {
	parsers.ErrCantContinue
}

// loaders errors
type ErrUnsupportedProtocol string

func (err ErrUnsupportedProtocol) Error() string {
	return fmt.Sprintf("available loaders can't work with protocol: %s", err)
}

type ErrDownloadingError string

func (err ErrDownloadingError) Error() string {
	return fmt.Sprintf("can't download: %s", err)
}
