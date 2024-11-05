package util

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/Francesco99975/qbittal/internal/models"
)

func QbittLogin(client *http.Client) error {
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")

	// Step 1: Sign In
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", qbittorrentAPI)
	loginData := url.Values{}
	loginData.Set("username", os.Getenv("QBITTORRENT_USERNAME"))
	loginData.Set("password", os.Getenv("QBITTORRENT_PASSWORD"))

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(loginData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request <- %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login <- %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code <- %d", resp.StatusCode)
	}

	return nil
}

func QbittAddTorrent(client *http.Client, torrent models.Torrent, downloadPath string) error {
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")
	addTorrentURL := fmt.Sprintf("%s/api/v2/torrents/add", qbittorrentAPI)
	addTorrentData := url.Values{}
	addTorrentData.Set("urls", torrent.MagnetLink)
	addTorrentData.Set("savepath", downloadPath)

	req, err := http.NewRequest("POST", addTorrentURL, bytes.NewBufferString(addTorrentData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create add torrent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add torrent <- %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add torrent <- Status Code: %d - Body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func QbittDeleteTorrent(client *http.Client, torrentHash string, dfStr string) error {
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")
	endpoint := fmt.Sprintf("%s/api/v2/torrents/delete", qbittorrentAPI)
	data := url.Values{}
	data.Set("hashes", torrentHash)
	data.Set("deleteFiles", dfStr)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete torrent: %w", err)

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete torrent: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func QbittLogout(client *http.Client) error {
	qbittorrentAPI := os.Getenv("QBITTORRENT_API")
	logoutURL := fmt.Sprintf("%s/api/v2/auth/logout", qbittorrentAPI)
	req, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request <- %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logout <- %w", err)

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status code: %d", resp.StatusCode)

	}

	return nil
}
