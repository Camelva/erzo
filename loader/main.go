package loader

import (
	"erzo/loader/ffmpeg"
	"erzo/types"
	"fmt"
	"log"
	"net/url"
)

var loaders []types.Loader

func init() {
	ffmpegConf := ffmpeg.GetConfig()
	if ffmpegConf == nil {
		log.Fatalln("FFmpeg is required for work.")
	}
	ffmpegConf.Available = true
	loaders = append(loaders, *ffmpegConf)
}

func Go(formats types.Formats) (string, error) {
	if len(formats) < 1 {
		return "", fmt.Errorf("not enough data")
	}
	for _, f := range formats {
		fileName, err := process(f)
		if err != nil {
			log.Println(err)
		}
		return fileName, nil
	}
	return "", fmt.Errorf("not found suitable loader")
}

func process(f types.Format) (string, error) {
	// TODO: add multiple parsers support
	u, err := url.Parse(f["url"])
	if err != nil {
		return "", err
	}
	for _, loader := range loaders {
		if !loader.Available {
			continue
		}
		if loader.Name == "ffmpeg" {
			songName, err := ffmpeg.Get(u)
			if err != nil {
				return "", err
			}
			return songName, nil
		}
	}
	return "", fmt.Errorf("not found suitable loader")
}