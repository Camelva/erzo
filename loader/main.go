package loader

import (
	"erzo/types"
	"log"
	"net/url"
)

func Go(formats types.Formats) string {
	if len(formats) < 1 {
		return "not enough data!"
	}
	for _, f := range formats {
		result := identify(f)
		if result == 1 {
			break
		} else {
			continue
		}
	}
	return "Okay"
}

func identify(format types.Format) (result int) {
	u, err := url.Parse(format["url"])
	if err != nil {
		log.Println(err)
		return 0
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		commonLoader(format)
	} else {
		log.Println("Need another loader")
		return 0
	}
	return 1
}