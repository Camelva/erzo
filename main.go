package erzo

import (
	"errors"
	"fmt"

	"github.com/camelva/erzo/engine"
	_ "github.com/camelva/erzo/loaders/ffmpeg"
	"github.com/camelva/erzo/parsers"
	_ "github.com/camelva/erzo/parsers/soundcloud"
)

//func main() {
//	reader := bufio.NewReader(os.Stdin)
//	fmt.Print("Enter link: ")
//	userInput, _ := reader.ReadString('\n')
//	r, err := Get(userInput)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	//_ = r
//	log.Println(r)
//}

var (
	ErrNotUrl               = errors.New("there is no valid url")
	ErrNotSupportedFormat   = errors.New("this format unsupported yet")
	ErrNotSupportedPlaylist = errors.New("playlists unsupported yet")
	ErrNotSupportedService  = errors.New("this service unsupported yet")
	ErrTryAgain             = errors.New("please try again")
)

type options struct {
	output   string
	truncate bool
	debug    bool
}

type Option interface {
	apply(*options)
}

type truncateOption bool

func (opt truncateOption) apply(opts *options) {
	opts.truncate = bool(opt)
}
func Truncate(b bool) Option {
	return truncateOption(b)
}

type outputOption string

func (opt outputOption) apply(opts *options) {
	opts.output = string(opt)
}
func Output(s string) Option {
	return outputOption(s)
}

type debugOption bool

func (opt debugOption) apply(opts *options) {
	opts.debug = true
}
func Debug(b bool) Option {
	return debugOption(b)
}

// Get process given url and download song from it.
// @message - url to process
// @options:
// 		Truncate(true|false) - clear output folder before processing
//		Output(string)		 - change output folder
//		Debug(true|false)    - log debug info
func Get(message string, opts ...Option) (string, error) {
	options := options{
		output:   "out",
		truncate: false,
		debug:    false,
	}
	for _, o := range opts {
		o.apply(&options)
	}
	e := engine.New(
		options.output,
		options.truncate,
		options.debug,
	)
	r, err := e.Process(message)
	if err != nil {
		if err == engine.ErrNotURL {
			return "", ErrNotUrl
		}
		if err, ok := err.(parsers.ErrNotSupported); ok {
			if err.Subject == "playlist" {
				return "", ErrNotSupportedPlaylist
			}
			return "", ErrNotSupportedFormat
		}
		if err.Error() == "unsupported service" {
			return "", ErrNotSupportedService
		}
		engine.Log("main", fmt.Errorf("can't process url `%s`. Error: %s", message, err))
		return "", ErrTryAgain
	}
	return r, nil
}
