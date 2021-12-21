package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/blang/semver/v4"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// APITimeout is how long we wait for the GitHub API
	APITimeout = 30 * time.Second

	// BaseURL is exported for tests
	BaseURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

// Asset is a GitHub release asset
type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is a GitHub release
type Release struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	TagName     string         `json:"tag_name"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []Asset        `json:"assets"`
	Version     semver.Version `json:"-"`
}

// FetchLatestLauncherRelease fetches meta-data about the latest release from GitHub
func FetchLatestLauncherRelease(ctx context.Context, githubOrg, githubRepo string) (Release, error) {

	ctx, cancel := context.WithTimeout(ctx, APITimeout)
	defer cancel()

	url := fmt.Sprintf(BaseURL, githubOrg, githubRepo)
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
		return rs, err
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

// return: bool exit
func UpdateLauncherFromNewBinary(ui *gui_win32.Gui, p *ipc_.Handler) bool {
	if utils.LauncherUpgradeAvailable() {
		update := ui.YesNoModal("Mysterium launcher upgrade", "You are running a newer version of launcher.\r\nUpgrade launcher installation ?")
		if model.IDYES == update {
			if !p.OwnsPipe() {
				p.SendStopApp()
				p.OpenPipe()
			}
			utils.UpdateExe()
			return false
		}
	}

	if !p.OwnsPipe() {
		p.SendPopupApp()
		return true
	}
	return false
}
