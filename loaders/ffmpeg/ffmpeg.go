package ffmpeg

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
)

var debugInstance = "ffmpeg"

type loader struct {
	name      string
	bin       string
	protocols []string
}

func (l loader) Name() string {
	return l.name
}
func (l loader) Bin() string {
	return l.bin
}
func (l loader) Get(u *url.URL, outName string) error {
	res, err := execute(
		l.Bin(),
		"-i",
		u.String(),
		"-c",
		"copy",
		outName,
		//"-report",
	)
	if err != nil {
		engine.Log(debugInstance, fmt.Errorf("can't execute command: %s", err))
		return err
	}
	engine.Log(debugInstance, fmt.Errorf("%s", res))
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

func init() {
	protocols := []string{"http", "https", "hls", "progressive"}
	bin := findBin()
	if len(bin) < 1 {
		engine.Log(debugInstance, fmt.Errorf("can't find ffmpeg binary"))
		return
	}
	config = loader{"ffmpeg", bin, protocols}
	engine.AddLoader(config)
}

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
