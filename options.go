package erzo

import "net/http"

type Options struct {
	Debug      bool
	Output     string
	Truncate   bool
	HTTPClient *http.Client
}

type Option func(*Options)

func Debug(debug bool) Option {
	return func(args *Options) {
		args.Debug = debug
	}
}

func Truncate(truncate bool) Option {
	return func(args *Options) {
		args.Truncate = truncate
	}
}

func Output(path string) Option {
	return func(args *Options) {
		args.Output = path
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(args *Options) {
		args.HTTPClient = client
	}
}
