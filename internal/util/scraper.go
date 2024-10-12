package util

import (
	"fmt"
	"net/http"
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

	form := url.Values{}
	form.Add("urls", torrent.MagnetLink)
	form.Add("savepath", pattern.DownloadPath)

	resp, err := http.PostForm(qbittorrentAPI, form)
	if err != nil {
		log.Errorf("Error sending request to qbittorrent API <-", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode == http.StatusOK {
		log.Info("Magnet link added successfully!")
	} else {
		log.Errorf("Failed to add magnet link: %s\n", resp.Status)
	}
}
