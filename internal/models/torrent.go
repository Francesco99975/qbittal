package models

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type DLTorrent struct {
	Hash       string
	Progress   float64
	Quit       chan struct{}
	HttpClient *http.Client
}

func NewDLTorrent(hash string) *DLTorrent {
	// Create a new HTTP client with a cookie jar to store cookies
	var jar, _ = cookiejar.New(nil)
	return &DLTorrent{
		Hash:       hash,
		Progress:   0.0,
		Quit:       make(chan struct{}),
		HttpClient: &http.Client{Jar: jar},
	}
}

func (t *DLTorrent) UpdateProgress(progress float64) {
	t.Progress = progress
}

type Torrent struct {
	MagnetLink string    `json:"magnet_link"`
	Keywords   []string  `json:"keywords"`
	Seeders    int       `json:"seeders"`
	Leechers   int       `json:"leechers"`
	Size       int       `json:"size"` // size in Mega Bytes
	Quality    int       `json:"quality"`
	Uploaded   time.Time `json:"uploaded"`
}

func (t *Torrent) IsValid() bool {
	return t.Seeders > 0 && t.Size > 0 && len(t.Keywords) > 0 && len(t.MagnetLink) > 0 && strings.Contains(t.MagnetLink, "magnet")
}

func (t *Torrent) CalculateQuality(SearchKeywords []string) {
	for _, sk := range SearchKeywords {
		for _, keyword := range t.Keywords {
			if strings.Contains(strings.ToLower(keyword), strings.ToLower(sk)) {
				t.Quality++
			}
		}
	}
}
