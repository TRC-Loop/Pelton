package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// update checking is off by default and, when on, only ever talks to GitHub's
// public releases API to compare tags - no telemetry, no other endpoint. The
// frequency setting controls how often the automatic (startup) check runs;
// the "check now" button in settings bypasses it and always checks.

const (
	settingUpdateCheckFreq = "update_check_frequency"
	settingLastUpdateCheck = "last_update_check_unix"

	defaultUpdateCheckFrequency = "off"

	updateCheckStartup = "startup"
	updateCheckWeekly  = "weekly"
	updateCheckMonthly = "monthly"

	latestReleaseAPI = "https://api.github.com/repos/TRC-Loop/Pelton/releases/latest"
)

// UpdateCheckResult is returned by a manual check and carried on
// EventUpdateAvailable for an automatic one.
type UpdateCheckResult struct {
	Checked        bool   `json:"checked"`
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	Error          string `json:"error"`
}

// githubRelease is the subset of GitHub's release API response we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates does an immediate GitHub releases check regardless of the
// update_check_frequency setting (the settings "Check now" button), and
// records it as the last check so the automatic schedule doesn't immediately
// re-check on next startup.
func (a *App) CheckForUpdates() (UpdateCheckResult, error) {
	if err := a.ready(); err != nil {
		return UpdateCheckResult{}, err
	}
	result := a.checkForUpdates(a.ctx)
	a.recordUpdateCheck()
	return result, nil
}

// maybeAutoCheckForUpdates runs the frequency-driven background check. It is
// called once from startup in its own goroutine so a slow or unreachable
// network never delays app startup.
func (a *App) maybeAutoCheckForUpdates(ctx context.Context) {
	freq := a.stringSetting(settingUpdateCheckFreq, defaultUpdateCheckFrequency)
	if freq == defaultUpdateCheckFrequency {
		return
	}
	if !a.updateCheckDue(freq) {
		return
	}

	result := a.checkForUpdates(ctx)
	a.recordUpdateCheck()
	if result.Checked {
		a.emit(EventUpdateAvailable, result)
	}
}

// updateCheckDue reports whether enough time has passed since the last check
// for the given frequency. "startup" is always due (every launch).
func (a *App) updateCheckDue(freq string) bool {
	if freq == updateCheckStartup {
		return true
	}
	last := a.intSetting(settingLastUpdateCheck, 0)
	if last == 0 {
		return true
	}
	elapsed := time.Since(time.Unix(int64(last), 0))
	switch freq {
	case updateCheckWeekly:
		return elapsed >= 7*24*time.Hour
	case updateCheckMonthly:
		return elapsed >= 30*24*time.Hour
	default:
		return false
	}
}

func (a *App) recordUpdateCheck() {
	_ = a.store.Set(a.ctx, settingLastUpdateCheck, strconv.FormatInt(time.Now().Unix(), 10))
}

// checkForUpdates does the actual GitHub API round trip and version compare.
// Errors come back in the result (not as a Go error) so a network hiccup
// shows as "couldn't check" in the ui rather than a hard failure.
func (a *App) checkForUpdates(ctx context.Context) UpdateCheckResult {
	result := UpdateCheckResult{CurrentVersion: a.version}

	release, err := fetchLatestRelease(ctx)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Checked = true
	result.LatestVersion = release.TagName
	result.ReleaseURL = release.HTMLURL
	result.Available = isVersionNewer(a.version, release.TagName)
	return result
}

func fetchLatestRelease(ctx context.Context) (githubRelease, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("pelton: github releases api returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, err
	}
	return release, nil
}

// isVersionNewer reports whether latest is a newer release than current.
// Both are compared as dot-separated numeric segments after stripping a
// leading "v" and any "-dirty"/"-N-gHASH" git-describe suffix. If current
// does not parse as a clean version (a local "dev" build, for instance),
// there is nothing meaningful to compare against, so it reports false rather
// than nagging a developer running an unreleased build.
func isVersionNewer(current, latest string) bool {
	c, ok := parseVersion(current)
	if !ok {
		return false
	}
	l, ok := parseVersion(latest)
	if !ok {
		return false
	}
	for i := 0; i < len(c) || i < len(l); i++ {
		var cv, lv int
		if i < len(c) {
			cv = c[i]
		}
		if i < len(l) {
			lv = l[i]
		}
		if lv != cv {
			return lv > cv
		}
	}
	return false
}

func parseVersion(v string) ([]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	// drop anything past the first "-" (git-describe suffixes like
	// "-3-gabcdef" or "-dirty", or a prerelease tag).
	if i := strings.IndexByte(v, '-'); i >= 0 {
		v = v[:i]
	}
	if v == "" {
		return nil, false
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}
