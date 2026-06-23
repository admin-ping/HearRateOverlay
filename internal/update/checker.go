package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Release represents a GitHub Release.
type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
}

// Checker handles checking for updates via GitHub API.
type Checker struct {
	owner      string
	repo       string
	currentVer string
	client     *http.Client
}

// NewChecker creates an update checker.
func NewChecker(owner, repo, currentVersion string) *Checker {
	return &Checker{
		owner:      owner,
		repo:       repo,
		currentVer: currentVersion,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// CheckLatest fetches the latest release and compares versions.
// Returns (latestRelease, hasUpdate, error).
func (c *Checker) CheckLatest() (*Release, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("decode failed: %w", err)
	}

	hasUpdate := compareVersions(release.TagName, c.currentVer) > 0
	return &release, hasUpdate, nil
}

// compareVersions compares two semantic version strings.
// Returns 1 if a > b, -1 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	// Strip leading 'v' if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB)
		}
		if numA > numB {
			return 1
		}
		if numA < numB {
			return -1
		}
	}
	return 0
}
