package types

import "time"

type ExtractorInfo struct {
	ID int
	Uploader string
	UploaderID int
	UploaderURL string
	Timestamp time.Time
	Title string
	Description string
	Thumbnails []Artwork
	Duration float32
	WebPageURL string
	License string
	ViewCount int
	LikeCount int
	CommentCount int
	RepostCount int
	Genre string
	Formats Formats
}

type Formats []Format
type Format map[string]string

type Artwork struct{
	Type string
	URL string
	Size int
}