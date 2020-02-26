package main

import "testing"

func TestGet(t *testing.T) {
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
			false,
		},
		{
			"Invalid url",
			args{"http:// and its all"},
			"",
			false,
		},
		{
			"SoundCloud only url",
			args{"https://soundcloud.com/whenzz/4-u"},
			"https://soundcloud.com/whenzz/4-u",
			false,
		},
		{
			"SoundCloud share message",
			args{`Listen to pain you feel (prod. rosesleeves) by rosesleeves on #SoundCloud
https://soundcloud.com/rosesleeves/painyoufeel`},
			"https://soundcloud.com/rosesleeves/painyoufeel",
			false,
		},
		{
			"SoundCloud embed player",
			args{`https://w.soundcloud.com/player/?visual=true&url=https%3A%2F%2Fapi.soundcloud.com%2Fplaylists%2F922213810&show_artwork=true&maxwidth=640&maxheight=960&dnt=1&secret_token=s-ziYey`},
			`https://w.soundcloud.com/player/?visual=true&url=https%3A%2F%2Fapi.soundcloud.com%2Fplaylists%2F922213810&show_artwork=true&maxwidth=640&maxheight=960&dnt=1&secret_token=s-ziYey`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.message)
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
