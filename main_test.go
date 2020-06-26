package erzo

import (
	"net/url"
	"path"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	var outFolder = "test_out"
	var trunc = true

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
			args{"https://soundcloud.com/nybillion/sneaky-cats-live-jam-session"},
			path.Join(outFolder, "sneaky-cats-live-jam-session.mp3"),
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
			got, meta, err := Get(tt.args.message, OptionOutput(outFolder), OptionTruncate(trunc))
			_ = meta
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

func TestGetDebug(t *testing.T) {
	var outFolder = "test_out"
	var trunc = true

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
			"SoundCloud only url",
			args{"https://soundcloud.com/user867574303/kurs-mind-detox-20-meditatsiya-1/s-fgHTh"},
			path.Join(outFolder, "kurs-mind-detox-20-meditatsiya-1.mp3"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, meta, err := Get(tt.args.message, OptionOutput(outFolder), OptionTruncate(trunc))
			_ = meta
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

func TestCheckURL(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		want    TidyURL
		wantErr bool
	}{
		{
			name:    "string",
			args:    args{msg: "some test string"},
			want:    TidyURL{},
			wantErr: true,
		},
		{
			name: "url",
			args: args{msg: "https://soundcloud.com/whenzz/4-u"},
			want: TidyURL{
				URL: url.URL{
					Scheme: "https",
					Host:   "soundcloud.com",
					Path:   "/whenzz/4-u",
				},
				Service: "SoundCloud",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckURL(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
