package ffmpeg

import (
	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
	"net/url"
	"os/exec"
)

type loader struct {
	name      string
	bin       string
	protocols []string
}

func init() {
	protocols := []string{"http", "https", "hls", "progressive"}
	bin := findBin()
	if len(bin) < 1 {
		// not found binary, so don't init loader
		return
	}
	config = loader{"ffmpeg", bin, protocols}
	engine.AddLoader(config)
}

func (l loader) Name() string {
	return l.name
}
func (l loader) Bin() string {
	return l.bin
}
func (l loader) Get(u *url.URL, outName string) error {
	_, err := execute(
		l.Bin(),
		"-y", // overwrite existing files
		"-i",
		u.String(),
		"-c",
		"copy",
		outName,
		//"-report", // only for debug
	)
	if err != nil {
		return err
	}
	return nil
}

func (l loader) Compatible(f parsers.Format) bool {
	for _, p := range l.protocols {
		if p != f.Protocol {
			continue
		}
		return true
	}
	return false
}

var config loader

func findBin() string {
	bin := "ffmpeg"
	path, err := exec.LookPath(bin)
	if err != nil {
		path = bin
	}
	return path
}

func execute(command ...string) ([]byte, error) {
	cmd := exec.Command(command[0], command[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}
