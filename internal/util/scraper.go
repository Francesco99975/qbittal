package util

import (
	"bytes"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Francesco99975/qbittal/internal/helpers"
	"github.com/Francesco99975/qbittal/internal/models"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gocolly/colly"
	"github.com/labstack/gommon/log"
)

var DownloadingTorrents = make(map[string]*models.DLTorrent, 0)
var Mu = &sync.RWMutex{}

func Scraper(pattern models.Pattern) {
	c := colly.NewCollector(colly.AllowURLRevisit(), colly.MaxDepth(100))
	torrents := make([]models.Torrent, 0)

	// Nyaa Scraper
	c.OnHTML("table.torrent-list>tbody", func(e *colly.HTMLElement) {
		if pattern.Source == models.Nyaa {

			e.ForEach("tr", func(i int, el *colly.HTMLElement) {

				magnetLink, err := nyaaMagentLinkFinder(el.ChildAttrs("td>a", "href"))
				if err != nil {
					log.Errorf("Error while finding magnet link <- %v", err)

				}

				log.Debugf("MagnetLink Found: %s", magnetLink)

				title, err := nyaaTitleFinder(el.ChildAttrs("td>a", "title"), pattern.SearchKeywords)
				if err != nil {
					log.Errorf("Error while finding title <- %v", err)
				}

				log.Debugf("Title Found: %s", title)

				size, err := evaluateSize(el.ChildText("td:nth-child(4)"))
				if err != nil {
					log.Errorf("Error while evaluating size <- %v", err)
				}

				log.Debugf("Size Found: %s", size)

				layout := "2006-01-02 15:04"
				uploadedRaw := el.ChildText("td:nth-child(5)")
				uploaded, err := time.Parse(layout, uploadedRaw)
				if err != nil {
					log.Errorf("Error while evaluating uploaded <- %v", err)
				}

				log.Debugf("Uploaded Found: %s", uploaded)

				seeders, err := strconv.Atoi(el.ChildText("td:nth-child(6)"))
				if err != nil {
					log.Errorf("Error while evaluating seeders <- %v", err)
				}

				log.Debugf("Seeders Found: %s", seeders)

				leechers, err := strconv.Atoi(el.ChildText("td:nth-child(7)"))
				if err != nil {
					log.Errorf("Error while evaluating leechers <- %v", err)
				}

				log.Debugf("Leechers Found: %s", leechers)

				torrent := models.Torrent{
					MagnetLink: magnetLink,
					Keywords:   strings.Split(title, " "),
					Size:       size,
					Seeders:    seeders,
					Leechers:   leechers,
					Uploaded:   uploaded,
				}

				torrent.CalculateQuality(pattern.SearchKeywords)

				torrents = append(torrents, torrent)
			})
		}
	})

	scrapingEndpoint := fmt.Sprintf("https://%s/%s=%s", pattern.Source, evaluateSearchUrl(pattern.Source), strings.Join(pattern.QueryKeywords, "+"))

	log.Debugf("Scraping %s", scrapingEndpoint)

	if pattern.Source == models.Nyaa {
		err := c.Visit(scrapingEndpoint)
		if err != nil {
			log.Errorf("Error while visiting the page <- %v", err)
		}
		c.Wait()
	} else if pattern.Source == models.PirateBay {
		// Run Rod in headless mode
		u := launcher.New().Headless(true).MustLaunch()
		browser := rod.New().ControlURL(u).MustConnect()

		// Open the target page
		page := browser.MustPage(scrapingEndpoint).MustWaitLoad()

		// Wait for the main list of torrents to load
		torrentItems := page.MustElements("ol#torrents li.list-entry")

		for _, el := range torrentItems {
			// Extract the magnet link
			potentialMagAttrs := el.MustElements("span.item-icons > a")
			var magnetLink string
			if len(potentialMagAttrs) > 0 {
				magnetLink = *potentialMagAttrs[0].MustAttribute("href")
				log.Printf("MagnetLink Found: %s", magnetLink)
			} else {
				log.Printf("MagnetLink not found")
				continue
			}

			// Extract the title
			title := strings.TrimSpace(el.MustElement("span.item-title > a").MustText())
			log.Printf("Title Found: %s", title)

			// Evaluate the size
			sizeText, err := strconv.Atoi(*el.MustElement("span.item-size > input").MustAttribute("value"))
			if err != nil {
				log.Printf("Error while evaluating size: %v", err)
			}

			size := bytesToMegabytes(sizeText)
			if err != nil {
				log.Printf("Error while evaluating size: %v", err)
			}
			log.Printf("Size Found: %s", size)

			// Parse the upload date
			layout := "2006-01-02"
			uploadedRaw := strings.TrimSpace(el.MustElement("span.item-uploaded > label").MustText())
			uploaded, err := time.Parse(layout, uploadedRaw)
			if err != nil {
				log.Printf("Error while evaluating uploaded: %v", err)
			}
			log.Printf("Uploaded Found: %s", uploaded)

			// Extract the number of seeders
			seedersText := strings.TrimSpace(el.MustElement("span.item-seed").MustText())
			seeders, err := strconv.Atoi(seedersText)
			if err != nil {
				log.Printf("Error while evaluating seeders: %v", err)
			}
			log.Printf("Seeders Found: %d", seeders)

			// Extract the number of leechers
			leechersText := strings.TrimSpace(el.MustElement("span.item-leech").MustText())
			leechers, err := strconv.Atoi(leechersText)
			if err != nil {
				log.Printf("Error while evaluating leechers: %v", err)
			}
			log.Printf("Leechers Found: %d", leechers)

			// Collect all data into a Torrent struct
			torrent := models.Torrent{
				MagnetLink: magnetLink,
				Keywords:   strings.Split(title, " "),
				Size:       size,
				Seeders:    seeders,
				Leechers:   leechers,
				Uploaded:   uploaded,
			}

			torrent.CalculateQuality(pattern.SearchKeywords)

			torrents = append(torrents, torrent)
		}

		browser.MustClose()
	}

	log.Debugf("Number of Torrents found: %d", len(torrents))
	prettyPrintTorrents(torrents)

	if len(torrents) == 0 {
		log.Errorf("No torrents found")
		return
	}

	// Filter out torrents with a lower thatn maximum quality
	max, err := findMaxQualityFromTorrents(torrents)
	if err != nil {
		log.Errorf("Error while finding max quality <- %v", err)
	}

	filteredTorrents := helpers.FilteredSlice(torrents, func(torrent models.Torrent) bool {
		return torrent.Quality >= max && torrent.Seeders > 5 && titleMatches(strings.Join(torrent.Keywords, " "), pattern.SearchKeywords)
	})

	log.Debugf("Number of Torrents after filtering: %d", len(filteredTorrents))
	prettyPrintTorrents(filteredTorrents)

	// Sort torrents by most seeders
	helpers.SortSlice(filteredTorrents, func(a, b models.Torrent) bool {
		return a.Seeders > b.Seeders
	})

	log.Debug("< < < Torrents after seeders sorting > > >")
	prettyPrintTorrents(filteredTorrents)

	//Sort torrents by most recent uploaded
	helpers.SortSlice(filteredTorrents, func(a, b models.Torrent) bool {
		return a.Uploaded.After(b.Uploaded)
	})

	log.Debug("< < < Torrents after date sorting  > > >")
	prettyPrintTorrents(filteredTorrents)

	// Download the top torrent by making a request to the qbittorrent API
	torrent := filteredTorrents[0]

	hash, err := extractHash(torrent.MagnetLink)
	if err != nil {
		log.Errorf("Error while extracting hash <- %v", err)
	}

	Mu.Lock()
	DownloadingTorrents[pattern.ID] = models.NewDLTorrent(hash)
	client := DownloadingTorrents[pattern.ID].HttpClient
	Mu.Unlock()

	log.Infof("Chosen torrent: %v", torrent)

	qbittorrentAPI := os.Getenv("QBITTORRENT_API")

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
		return
	}

	// Step 2: Add Torrent
	addTorrentURL := fmt.Sprintf("%s/api/v2/torrents/add", qbittorrentAPI)
	addTorrentData := url.Values{}
	addTorrentData.Set("urls", torrent.MagnetLink)
	addTorrentData.Set("savepath", pattern.DownloadPath)

	req, err = http.NewRequest("POST", addTorrentURL, bytes.NewBufferString(addTorrentData.Encode()))
	if err != nil {
		log.Errorf("failed to create add torrent request: %w", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("failed to add torrent: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("failed to add torrent: %d - %s", resp.StatusCode, string(body))
		return
	}

	go trackProgress(pattern.ID)

	log.Infof("Torrent added successfully")
}

func evaluateSearchUrl(source models.Source) string {
	switch source {
	case models.Nyaa:
		return "?q"
	case models.PirateBay:
		return "search.php?q"
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

func titleMatches(title string, searchQuery []string) bool {
	for _, keyword := range searchQuery {
		if !strings.Contains(title, keyword) {
			return false
		}
	}
	return true
}

func prettyPrintTorrents(torrents []models.Torrent) {
	for _, torrent := range torrents {
		log.Debugf("Keywords: %s", strings.Join(torrent.Keywords, ", "))
		log.Debugf("Quality: %s", torrent.Quality)
		log.Debugf("Seeders: %d", torrent.Seeders)
		log.Debugf("Size: %s", torrent.Size)
		log.Debugf("Leechers: %d", torrent.Leechers)
		log.Debugf("Uploaded: %s", torrent.Uploaded)
		log.Debugf("----------------------------------------")
	}
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

// Convert bytes to megabytes
func bytesToMegabytes(bytes int) int {
	return bytes / 1024 / 1024
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

// extractHash extracts the torrent hash from a magnet link
func extractHash(magnetLink string) (string, error) {
	u, err := url.Parse(magnetLink)
	if err != nil {
		return "", fmt.Errorf("failed to parse magnet link: %v", err)
	}

	// Magnet link parameters are in the "xt" query parameter
	for _, param := range u.Query()["xt"] {
		if strings.HasPrefix(param, "urn:btih:") {
			return strings.TrimPrefix(param, "urn:btih:"), nil
		}
	}

	return "", fmt.Errorf("no valid hash found in magnet link")
}

// Track torrent progress
func trackProgress(id string) {
	Mu.RLock()
	torrent, exists := DownloadingTorrents[id]
	if !exists {
		Mu.RUnlock()
		return
	}
	Mu.RUnlock()
	for {
		select {
		case <-torrent.Quit:
			log.Debugf("Quitting torrent %s", torrent.Hash)
			return
		default:
			progress, isComplete, err := getTorrentProgress(torrent.Hash, torrent.HttpClient)
			if err != nil {
				log.Errorf("Failed to get torrent progress <- %v", err)
				return
			}
			if isComplete {
				log.Infof("Torrent %s completed", torrent.Hash)
				DeleteTorrent(id, false)
				return
			}
			log.Debugf("Torrent %s progress update to %f", torrent.Hash, progress)
			Mu.Lock()
			DownloadingTorrents[id].UpdateProgress(progress)
			Mu.Unlock()
			time.Sleep(2 * time.Second) // Poll every 2 seconds
		}
	}
}

// Get progress for specific torrent
func getTorrentProgress(torrentHash string, client *http.Client) (float64, bool, error) {
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")
	url := fmt.Sprintf("%s/api/v2/torrents/info?hashes=%s", qbittorrentAPI, torrentHash)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0.0, false, fmt.Errorf("failed to create request <- %v", err)
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return 0.0, false, fmt.Errorf("request failed with status %d for hash %s <- %v", resp.StatusCode, torrentHash, err)
	}
	defer resp.Body.Close()

	var torrents []map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, false, fmt.Errorf("failed to read response body <- %v", err)
	}
	err = json.Unmarshal(body, &torrents)
	if err != nil {
		return 0.0, false, fmt.Errorf("failed to unmarshal response body <- %v", err)
	}
	if len(torrents) == 0 {
		return 0.0, false, fmt.Errorf("no downloading torrents found for hash %s", torrentHash)
	}

	log.Debugf("LEN Torrents: %d", len(torrents))
	log.Debugf("Torrent Hash: %s", torrents[0]["hash"])
	log.Debugf("Torrent State: %s", torrents[0]["state"])
	log.Debugf("Torrent Progress: %f", torrents[0]["progress"])

	progress := torrents[0]["progress"].(float64)
	isComplete := torrents[0]["state"] == "stalledUP" || torrents[0]["state"] == "pausedUP"
	return progress, isComplete, nil
}

// Delete torrent after completion
func DeleteTorrent(id string, deleteFiles bool) {
	Mu.RLock()
	torrentHash := DownloadingTorrents[id].Hash
	client := DownloadingTorrents[id].HttpClient
	Mu.RUnlock()
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")
	dfStr := "false"
	if deleteFiles {
		dfStr = "true"
	}

	endpoint := fmt.Sprintf("%s/api/v2/torrents/delete", qbittorrentAPI)
	data := url.Values{}
	data.Set("hashes", torrentHash)
	data.Set("deleteFiles", dfStr)
	// url := fmt.Sprintf("%s/api/v2/torrents/delete?hashes=%s&deleteFiles=%s", qbittorrentAPI, torrentHash, dfStr)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to delete torrent: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("failed to delete torrent: %d - %s", resp.StatusCode, string(body))
	}

	// Step 3: Log Out
	logoutURL := fmt.Sprintf("%s/api/v2/auth/logout", qbittorrentAPI)
	req, err = http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		log.Errorf("failed to create logout request: %w", err)
		return
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("failed to logout: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("logout failed with status code: %d", resp.StatusCode)
		return
	}

	Mu.Lock()
	close(DownloadingTorrents[id].Quit)
	delete(DownloadingTorrents, id)
	Mu.Unlock()
}
