package updates

import (
	"time"

	"github.com/blang/semver/v4"
)

// Asset is a GitHub release asset
type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is a GitHub release
type Release struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	TagName     string         `json:"tag_name"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []Asset        `json:"assets"`
	Version     semver.Version `json:"-"`
}
