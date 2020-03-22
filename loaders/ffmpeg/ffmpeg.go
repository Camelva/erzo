package ffmpeg

import (
	"bytes"
	"erzo/engine"
	"erzo/parsers"
	"fmt"
	"net/url"
	"os/exec"
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
	_, err := execute(
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
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return ""
	}
	return path
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
