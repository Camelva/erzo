package ffmpeg

import (
	"bytes"
	"erzo/types"
	"log"
	"net/url"
	"os/exec"
)

var config types.Loader

func init() {
	bin, err := findBin()
	if err != nil {
		log.Println(err)
	}
	config = types.Loader{
		Name:    "ffmpeg",
		Bin:     bin,
		Formats: []string{"http"},
	}
}

func findBin() (string, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", err
	}
	return path, nil
}

func execute(command ...string) (bytes.Buffer, error) {
	var out bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return out, err
	}
	return out, nil
}

//func Version() (string, error) {
//	if config.Bin == "" {
//		return "", fmt.Errorf("can't find ffmpeg")
//	}
//	res, err := execute(config.Bin, "-version")
//	if err != nil {
//		return "", err
//	}
//	return res.String(), nil
//}

func GetConfig() *types.Loader {
	return &config
}

func Get(u *url.URL) (string, error) {
	var outName = "song.mp3"
	_, err := execute(config.Bin, "-i", u.String(), "-c", "copy", outName)
	if err != nil {
		return "", err
	}
	return outName, nil
}
