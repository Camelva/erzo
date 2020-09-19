package erzo

import (
	"github.com/camelva/erzo/engine"
	_ "github.com/camelva/erzo/loaders/ffmpeg"
	_ "github.com/camelva/erzo/parsers/soundcloud"
	_ "github.com/camelva/erzo/parsers/youtube"
)

// Get process given url and download song from it.
// @message - url to process
// @options:
// 		Truncate(true|false) - clear output folder before processing
//		Output(string)		 - change output folder
// Return filename or one of the following errors:
// ErrNotURL if there is no urls in your message
// ErrUnsupportedService if url belongs to unsupported service
// ErrUnsupportedType if service supported but certain type - not yet
// ErrCantFetchInfo if fatal error occurred while extracting info from url
// ErrUnsupportedProtocol if there is no downloader for this format
// ErrDownloadingError if fatal error occurred while downloading song
// ErrUndefined any other errors
func Get(message string, setters ...Option) (*engine.SongResult, error) {
	song, err := GetInfo(message, setters...)
	if err != nil {
		return nil, err
	}
	songRes, err := song.Get()
	if err != nil {
		return nil, convertErr(err)
	}
	return songRes, nil
}

func GetInfo(message string, setters ...Option) (*engine.SongInfo, error) {
	args := Options{
		Debug:      false,
		Output:     "out",
		Truncate:   false,
		HTTPClient: nil,
	}
	for _, setter := range setters {
		setter(&args)
	}
	e := engine.New(
		args.Output,
		args.Truncate,
		args.Debug,
		args.HTTPClient,
	)
	song, err := e.GetInfo(message)
	if err != nil {
		return nil, convertErr(err)
	}
	return song, nil
}

func convertErr(err error) error {
	var convertedErr error
	switch err.(type) {
	case engine.ErrNotURL:
		convertedErr = ErrNotURL{err.(engine.ErrNotURL)}
	case engine.ErrUnsupportedService:
		convertedErr = ErrUnsupportedService{err.(engine.ErrUnsupportedService)}
	case engine.ErrUnsupportedType:
		convertedErr = ErrUnsupportedType{err.(engine.ErrUnsupportedType)}
	case engine.ErrCantFetchInfo:
		convertedErr = ErrCantFetchInfo{err.(engine.ErrCantFetchInfo)}
	case engine.ErrUnsupportedProtocol:
		convertedErr = ErrUnsupportedProtocol{err.(engine.ErrUnsupportedProtocol)}
	case engine.ErrDownloadingError:
		convertedErr = ErrDownloadingError{err.(engine.ErrDownloadingError)}
	case engine.ErrUndefined:
		convertedErr = ErrUndefined{err.(engine.ErrUndefined)}
	default:
		convertedErr = ErrUndefined{engine.ErrUndefined{}}
	}
	return convertedErr
}
