package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/kurlmarx/romwrangler/internal/systems"
)

const (
	screenScraperBaseURL = "https://api.screenscraper.fr/api2"
	softwareName         = "RomWrangler"
)

// ScreenScraperClient accesses the ScreenScraper API v2.
type ScreenScraperClient struct {
	devID    string
	devPass  string
	user     string
	password string
	client   *http.Client

	mu       sync.Mutex
	lastCall time.Time
}

// NewScreenScraperClient creates a new ScreenScraper API client.
func NewScreenScraperClient(user, password string) *ScreenScraperClient {
	return &ScreenScraperClient{
		devID:    "romwrangler",
		devPass:  "",
		user:     user,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ssGameInfo is the ScreenScraper API response structure (simplified).
type ssResponse struct {
	Header struct {
		Success string `json:"success"`
	} `json:"header"`
	Response struct {
		Game ssGame `json:"jeu"`
	} `json:"response"`
}

type ssGame struct {
	ID       int      `json:"id"`
	Names    []ssName `json:"noms"`
	SystemID int      `json:"systemeid"`
	Dates    []ssDate `json:"dates"`
	Publisher struct {
		Text string `json:"text"`
	} `json:"editeur"`
	Synopsis []ssSynopsis `json:"synopsis"`
}

type ssName struct {
	Region string `json:"region"`
	Text   string `json:"text"`
}

type ssDate struct {
	Region string `json:"region"`
	Text   string `json:"text"`
}

type ssSynopsis struct {
	Language string `json:"langue"`
	Text     string `json:"text"`
}

// Identify tries to identify a ROM by its hashes via the ScreenScraper API.
func (c *ScreenScraperClient) Identify(ctx context.Context, hashes FileHashes, systemID systems.SystemID) (*GameInfo, error) {
	c.rateLimit()

	params := url.Values{}
	params.Set("devid", c.devID)
	params.Set("devpassword", c.devPass)
	params.Set("softname", softwareName)
	params.Set("output", "json")

	if c.user != "" {
		params.Set("ssid", c.user)
		params.Set("sspassword", c.password)
	}

	if hashes.CRC32 != "" {
		params.Set("crc", hashes.CRC32)
	}
	if hashes.MD5 != "" {
		params.Set("md5", hashes.MD5)
	}
	if hashes.SHA1 != "" {
		params.Set("sha1", hashes.SHA1)
	}
	params.Set("romtaille", fmt.Sprintf("%d", hashes.Size))

	reqURL := screenScraperBaseURL + "/jeuInfos.php?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("screenscraper request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // not found
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("screenscraper returned %d: %s", resp.StatusCode, string(body))
	}

	var ssResp ssResponse
	if err := json.NewDecoder(resp.Body).Decode(&ssResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	game := ssResp.Response.Game

	// Get name (prefer world/us/eu regions)
	name := extractName(game.Names)
	year := extractDate(game.Dates)
	desc := extractSynopsis(game.Synopsis)

	// Use API-detected system if no systemID was provided
	detectedSystem := systemID
	if detectedSystem == "" {
		if mapped, ok := ScreenScraperToSystemID(game.SystemID); ok {
			detectedSystem = mapped
		}
	}

	return &GameInfo{
		Name:        name,
		System:      detectedSystem,
		Description: desc,
		Publisher:   game.Publisher.Text,
		Year:        year,
		Source:      "screenscraper",
	}, nil
}

func (c *ScreenScraperClient) rateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.lastCall)
	if elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}
	c.lastCall = time.Now()
}

func extractName(names []ssName) string {
	preferredRegions := []string{"wor", "us", "eu", "ss", "jp"}
	for _, region := range preferredRegions {
		for _, n := range names {
			if n.Region == region {
				return n.Text
			}
		}
	}
	if len(names) > 0 {
		return names[0].Text
	}
	return ""
}

func extractDate(dates []ssDate) string {
	preferredRegions := []string{"wor", "us", "eu", "jp"}
	for _, region := range preferredRegions {
		for _, d := range dates {
			if d.Region == region {
				return d.Text
			}
		}
	}
	if len(dates) > 0 {
		return dates[0].Text
	}
	return ""
}

func extractSynopsis(synopses []ssSynopsis) string {
	preferredLangs := []string{"en", "us"}
	for _, lang := range preferredLangs {
		for _, s := range synopses {
			if s.Language == lang {
				return s.Text
			}
		}
	}
	if len(synopses) > 0 {
		return synopses[0].Text
	}
	return ""
}
