package util

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"
	"github.com/gocolly/colly"
	"github.com/labstack/gommon/log"
)

func evaluateSearchUrl(source models.Source) string {
	switch source {
	case models.Nyaa:
		return "?q="
	case models.PirateBay:
		return "search.php?q="
	default:
		return ""
	}
}

func nyaaMagentLinkFinder(links []string) (string, error) {
	for _, link := range links {
		if strings.Contains(link, "magnet") {
			return link, nil
		}
	}
	return "", fmt.Errorf("no magnet link found")
}

func nyaaTitleFinder(titles []string, searchQuery []string) (string, error) {
	for _, title := range titles {
		for _, query := range searchQuery {
			if strings.Contains(title, query) {
				return title, nil
			}
		}
	}
	return "", fmt.Errorf("no title found")
}

func evaluateSize(size string) (int, error) {
	// Regular expression to match the number and the unit
	re := regexp.MustCompile(`([\d.]+)\s*(\w+)`)
	matches := re.FindStringSubmatch(size)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	// Parse the numeric part
	numberStr := matches[1]
	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, err
	}

	// Determine the unit and conversion factor
	unit := strings.ToLower(matches[2])
	var multiplier float64

	switch unit {
	case "mb", "mib":
		multiplier = 1 // already in MB
	case "gb":
		multiplier = 1000 // convert GB to MB
	case "gib":
		multiplier = 1024 // convert GiB to MB
	case "kb":
		multiplier = 1 / 1000.0 // convert KB to MB
	case "kib":
		multiplier = 1 / 1024.0 // convert KiB to MB
	case "tb":
		multiplier = 1000 * 1000 // convert TB to MB
	case "tib":
		multiplier = 1024 * 1024 // convert TiB to MB
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	// Convert to megabytes
	megabytes := number * multiplier

	return int(megabytes), nil

}

func findMaxQualityFromTorrents(torrents []models.Torrent) (int, error) {
	maxQuality := torrents[0].Quality
	for _, torrent := range torrents {
		if torrent.Quality > maxQuality {
			maxQuality = torrent.Quality
		}
	}
	return maxQuality, nil
}

func Scraper(pattern models.Pattern) {
	c := colly.NewCollector(colly.AllowURLRevisit(), colly.MaxDepth(100))
	torrents := make([]models.Torrent, 0)

	c.OnHTML("table.torrent-list>tbody", func(e *colly.HTMLElement) {
		if pattern.Source == models.Nyaa {
			e.ForEach("tr", func(i int, el *colly.HTMLElement) {

				magnetLink, err := nyaaMagentLinkFinder(el.ChildAttrs("td>a", "href"))

				if err != nil {
					log.Errorf("Error while finding magnet link <- %v", err)
				}

				title, err := nyaaTitleFinder(el.ChildAttrs("td>a", "title"), pattern.SearchKeywords)
				if err != nil {
					log.Errorf("Error while finding title <- %v", err)
				}

				size, err := evaluateSize(el.ChildText("td:nth-child(4)"))
				if err != nil {
					log.Errorf("Error while evaluating size <- %v", err)
				}

				seeders, err := strconv.Atoi(el.ChildText("td:nth-child(6)"))
				if err != nil {
					log.Errorf("Error while evaluating seeders <- %v", err)
				}

				leechers, err := strconv.Atoi(el.ChildText("td:nth-child(7)"))
				if err != nil {
					log.Errorf("Error while evaluating leechers <- %v", err)
				}

				torrent := models.Torrent{
					MagnetLink: magnetLink,
					Keywords:   strings.Split(title, " "),
					Size:       size,
					Seeders:    seeders,
					Leechers:   leechers,
				}

				torrent.CalculateQuality(pattern.SearchKeywords)

				torrents = append(torrents, torrent)
			})
		}
	})

	err := c.Visit(fmt.Sprintf("https://%s/%s=%s", pattern.Source, evaluateSearchUrl(pattern.Source), strings.Join(pattern.QueryKeywords, "+")))
	if err != nil {
		log.Errorf("Error while visiting the page <- %v", err)
	}

	c.Wait()

	// Filter out torrents with a lower thatn maximum quality
	max, err := findMaxQualityFromTorrents(torrents)
	if err != nil {
		log.Errorf("Error while finding max quality <- %v", err)
	}

	filteredTorrents := helpers.FilteredSlice(torrents, func(torrent models.Torrent) bool {
		return torrent.Quality >= max
	})

	// Sort torrents by most seeders
	helpers.SortSlice(filteredTorrents, func(a, b models.Torrent) bool {
		return a.Seeders > b.Seeders
	})

	// Download the top torrent by making a request to the qbittorrent API
	torrent := filteredTorrents[0]

	qbittorrentAPI := os.Getenv("QBITTORRENT_API")

	// Create a new HTTP client with a cookie jar to store cookies
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Step 1: Sign In
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", qbittorrentAPI)
	loginData := url.Values{}
	loginData.Set("username", os.Getenv("QBITTORRENT_USERNAME"))
	loginData.Set("password", os.Getenv("QBITTORRENT_PASSWORD"))

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(loginData.Encode()))
	if err != nil {
		log.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// Step 2: Add Torrent
	addTorrentURL := fmt.Sprintf("%s/api/v2/torrents/add", qbittorrentAPI)
	addTorrentData := url.Values{}
	addTorrentData.Set("urls", torrent.MagnetLink)
	addTorrentData.Set("savepath", pattern.DownloadPath)

	req, err = http.NewRequest("POST", addTorrentURL, bytes.NewBufferString(addTorrentData.Encode()))
	if err != nil {
		log.Errorf("failed to create add torrent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("failed to add torrent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("failed to add torrent: %d - %s", resp.StatusCode, string(body))
	}

	// Step 3: Log Out
	logoutURL := fmt.Sprintf("%s/api/v2/auth/logout", qbittorrentAPI)
	req, err = http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		log.Errorf("failed to create logout request: %w", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("failed to logout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("logout failed with status code: %d", resp.StatusCode)
	}
}
