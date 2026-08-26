package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/version"
)

const (
	githubReleasesURL = "https://api.github.com/repos/deranjer/loopira/releases/latest"
	latestCacheTTL    = 1 * time.Hour
)

var latestCache struct {
	mu      sync.Mutex
	tag     string
	fetched time.Time
}

type versionBody struct {
	Version         string `json:"version"`
	LatestVersion   string `json:"latestVersion,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
}

type versionOutput struct {
	Body versionBody
}

func (s *Server) registerVersionRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-version",
		Method:      http.MethodGet,
		Path:        "/api/v1/version",
		Summary:     "Server version and GitHub release check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*versionOutput, error) {
		out := &versionOutput{Body: versionBody{Version: version.Version}}
		if latest := latestRelease(ctx); latest != "" {
			out.Body.LatestVersion = latest
			out.Body.ReleaseURL = "https://github.com/deranjer/loopira/releases/tag/" + latest
			out.Body.UpdateAvailable = isNewer(latest, version.Version)
		}
		return out, nil
	})
}

// latestRelease returns the latest release tag from GitHub, cached for an
// hour so every navbar render doesn't hit the GitHub API. Returns "" if
// nothing has ever been fetched successfully; an unreachable GitHub
// shouldn't degrade the app, so failures just fall back to the stale value.
func latestRelease(ctx context.Context) string {
	latestCache.mu.Lock()
	if time.Since(latestCache.fetched) < latestCacheTTL {
		tag := latestCache.tag
		latestCache.mu.Unlock()
		return tag
	}
	latestCache.mu.Unlock()

	if tag, err := fetchLatestReleaseTag(ctx); err == nil {
		latestCache.mu.Lock()
		latestCache.tag = tag
		latestCache.fetched = time.Now()
		latestCache.mu.Unlock()
	}

	latestCache.mu.Lock()
	defer latestCache.mu.Unlock()
	return latestCache.tag
}

func fetchLatestReleaseTag(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubReleasesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.TagName, nil
}

// isNewer reports whether latest (e.g. "v1.4.2") is a greater semver than
// current (e.g. "v1.4.1" or "dev"). A non-semver current build (dev builds)
// is always considered behind a real release.
func isNewer(latest, current string) bool {
	l, ok := parseSemver(latest)
	if !ok {
		return false
	}
	c, ok := parseSemver(current)
	if !ok {
		return true
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parseSemver(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
