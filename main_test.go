package erzo

import (
	"path"
	"testing"
)

func TestGet(t *testing.T) {
	var outFolder = "test_out"
	var trunc = false

	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Not a url",
			args{"just some text"},
			"",
			true,
		},
		{
			"Invalid url",
			args{"http:// and its all"},
			"",
			true,
		},
		{
			"SoundCloud only url",
			args{"https://soundcloud.com/whenzz/4-u"},
			path.Join(outFolder, "4-u.mp3"),
			false,
		},
		{
			"Another song",
			args{"https://soundcloud.com/erynmartin/homie-wonderland-1"},
			path.Join(outFolder, "homie-wonderland-1.mp3"),
			false,
		},
		{
			"SoundCloud only url #2",
			args{"https://soundcloud.com/ynwmelly/suicidal-ft-juice-wrld-remix"},
			path.Join(outFolder, "suicidal-ft-juice-wrld-remix.mp3"),
			false,
		},
		{
			"SoundCloud share from album",
			args{"https://soundcloud.com/iamtrevordaniel/falling?in=iamtrevordaniel/sets/homesick"},
			path.Join(outFolder, "falling.mp3"),
			false,
		},
		{
			"SoundCloud removed track",
			args{"https://soundcloud.com/rosesleeves/painyoufeel"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.message, Output(outFolder), Truncate(trunc))
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
