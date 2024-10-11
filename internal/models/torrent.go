package models

type Torrent struct {
	MagnetLink string   `json:"magnet_link"`
	Keywords   []string `json:"keywords"`
	Seeders    int      `json:"seeders"`
	Leechers   int      `json:"leechers"`
	Size       int      `json:"size"` // size in Mega Bytes
	Quality    int      `json:"quality"`
}

func containsKeyword(keywords []string, keyword string) bool {
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}

func (t *Torrent) CalculateQuality(SearchKeywords []string) {
	for _, keyword := range SearchKeywords {
		if containsKeyword(t.Keywords, keyword) {
			t.Quality++
		}
	}
}
