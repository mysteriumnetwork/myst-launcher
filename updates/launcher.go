package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/blang/semver/v4"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// apiTimeout is how long we wait for the GitHub API
	apiTimeout = 30 * time.Second

	// baseURL is exported for tests
	baseURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

// FetchLatestRelease fetches meta-data about the latest release from GitHub
func FetchLatestRelease(ctx context.Context, githubOrg, githubRepo string) (Release, error) {
	log.Println("FetchLatestRelease>")

	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	url := fmt.Sprintf(baseURL, githubOrg, githubRepo)
	log.Println("FetchLatestRelease>", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, nil
	}

	// pin to API version 3 to avoid breaking our structs
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("request faild with %v (%v)", resp.StatusCode, resp.Status)
	}

	var rs Release
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return Release{}, err
	}

	//if !strings.HasPrefix(rs.TagName, "v") {
	//	return rs, fmt.Errorf("tag name %q is invalid, must start with 'v'", rs.TagName)
	//}
	v, err := semver.Parse(rs.TagName[0:])
	if err != nil {
		return rs, fmt.Errorf("failed to parse version %q: %q", rs.TagName[0:], err)
	}
	rs.Version = v

	return rs, nil
}
