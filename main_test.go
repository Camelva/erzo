package erzo

import (
	"github.com/camelva/erzo/engine"
	"path"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var outFolder = "test_out"
	var trunc = true
	var dateLayout = "2006-01-02"

	dur3, _ := time.ParseDuration("2m32s")
	date3, _ := time.Parse(dateLayout, "2020-02-23")

	dur4, _ := time.ParseDuration("4m14s")
	date4, _ := time.Parse(dateLayout, "2014-08-01")

	dur5, _ := time.ParseDuration("2m39s")
	date5, _ := time.Parse(dateLayout, "2018-10-03")

	dur7, _ := time.ParseDuration("2m50s")
	date7, _ := time.Parse(dateLayout, "2020-04-02")

	dur8, _ := time.ParseDuration("3m28s")
	date8, _ := time.Parse(dateLayout, "2020-08-12")

	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    *engine.SongResult
		wantErr bool
	}{
		{
			"#1 Not a url",
			args{"just some text"},
			nil,
			true,
		},
		{
			"#2 Invalid url",
			args{"http:// and its all"},
			nil,
			true,
		},
		{
			"#3 SoundCloud only url",
			args{"https://soundcloud.com/whenzz/4-u"},
			&engine.SongResult{
				Path:       path.Join(outFolder, "4-u.mp3"),
				Author:     "WheNzz 神",
				Title:      "4 U 💔",
				Thumbnails: nil,
				Duration:   dur3,
				UploadDate: date3,
			},
			false,
		},
		{
			"#4 Another song",
			args{"https://soundcloud.com/nybillion/sneaky-cats-live-jam-session"},
			&engine.SongResult{
				Path:       path.Join(outFolder, "sneaky-cats-live-jam-session.mp3"),
				Author:     "nybillion",
				Title:      "Sneaky Cats",
				Thumbnails: nil,
				Duration:   dur4,
				UploadDate: date4,
			},
			false,
		},
		{
			"#5 SoundCloud share from album",
			args{"https://soundcloud.com/iamtrevordaniel/falling?in=iamtrevordaniel/sets/homesick"},
			&engine.SongResult{
				Path:       path.Join(outFolder, "falling.mp3"),
				Author:     "Trevor Daniel",
				Title:      "Falling",
				Thumbnails: nil,
				Duration:   dur5,
				UploadDate: date5,
			},
			false,
		},
		{
			"#6 SoundCloud removed track",
			args{"https://soundcloud.com/rosesleeves/painyoufeel"},
			nil,
			true,
		},
		{
			name: "#7 Youtube",
			args: args{"https://www.youtube.com/watch?v=x9xJ4r6Vhnk&feature=youtu.be"},
			want: &engine.SongResult{
				Path:       path.Join(outFolder, ".ethereal - ASTRAL (ft. OSIAS).mp3"),
				Author:     "Bass Nation",
				Title:      ".ethereal - ASTRAL (ft. OSIAS)",
				Thumbnails: nil,
				Duration:   dur7,
				UploadDate: date7,
			},
			wantErr: false,
		},
		{
			"#8 Youtube",
			args{"https://youtu.be/fsjANtamXl4"},
			&engine.SongResult{
				Path:       path.Join(outFolder, "ÆSTRAL - reasons.mp3"),
				Author:     "Bass Nation",
				Title:      "ÆSTRAL - reasons",
				Thumbnails: nil,
				Duration:   dur8,
				UploadDate: date8,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.message, OptionOutput(outFolder), OptionTruncate(trunc))
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want != got {
					t.Errorf("Get() got %v", got)
				}
				return
			}
			if got.Path != tt.want.Path {
				t.Errorf("Get() got = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Get() got = %v, want %v", got.Title, tt.want.Title)
			}
			if got.Author != tt.want.Author {
				t.Errorf("Get() got = %v, want %v", got.Author, tt.want.Author)
			}
			if got.Duration != tt.want.Duration {
				t.Errorf("Get() got = %v, want %v", got.Duration, tt.want.Duration)
			}
			if got.UploadDate != tt.want.UploadDate {
				t.Errorf("Get() got = %v, want %v", got.UploadDate, tt.want.UploadDate)
			}
		})
	}
}

func TestGetInfoDebug(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{"first", "https://youtu.be/CPmuDfD8VI8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetInfo(tt.message, OptionOutput("debug"), OptionTruncate(true))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetInfo() result = %#v", got)
		})
	}
}

func TestGetDebug(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{"first", "https://youtu.be/CPmuDfD8VI8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.message, OptionOutput("debug"), OptionTruncate(true))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetInfo() result = %#v", got)
		})
	}
}
