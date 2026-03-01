package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const releasesRepo = "claraverse-space/clara-companion-releases"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatestVersion fetches the latest release tag from GitHub.
// Returns the version string (without leading "v") or an error.
func CheckLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", releasesRepo)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// IsNewer returns true if latest is a higher semver than current.
// Both should be in "X.Y.Z" format (no "v" prefix).
func IsNewer(current, latest string) bool {
	curParts := parseSemver(current)
	latParts := parseSemver(latest)
	if curParts == nil || latParts == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if latParts[i] > curParts[i] {
			return true
		}
		if latParts[i] < curParts[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	// Strip any suffix after a hyphen (e.g. "1.0.0-dev")
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}
