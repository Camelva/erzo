package parsers

import "fmt"

type ErrFormatNotSupported string

func (e ErrFormatNotSupported) Error() string {
	return fmt.Sprintf("not supported format: %s", e)
}

type ErrCantContinue string

func (e ErrCantContinue) Error() string {
	return fmt.Sprintf("process interrupted: %s", e)
}

type ErrURLParse struct {
	Message string
	URL     string
}

func (err ErrURLParse) Error() string {
	return fmt.Sprintf("url.Parse() %s failed: %s", err.URL, err.Message)
}
