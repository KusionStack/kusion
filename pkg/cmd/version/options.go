package version

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/djherbis/times"

	"kusionstack.io/kusion/pkg/clipath"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/gitutil"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/version"
)

const jsonOutput = "json"

type VersionOptions struct {
	Output string
}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func (o *VersionOptions) Validate() error {
	if o.Output != "" && o.Output != jsonOutput {
		return errors.New("invalid output type, output must be 'json'")
	}
	return nil
}

func (o *VersionOptions) Run() {
	if strings.ToLower(o.Output) == jsonOutput {
		fmt.Println(version.JSON())
	} else {
		fmt.Println(version.String())
		if msg := checkForUpdate(); msg != "" {
			fmt.Println(msg)
		}
	}
}

// checkForUpdate checks to see if the CLI needs to be updated,
// and if so emits a warning, as well as information as to how it can be upgraded.
func checkForUpdate() string {
	curVer, err := semver.ParseTolerant(version.ReleaseVersion())
	if err != nil {
		log.Errorf("error parsing current version: %s", err)
	}

	// We don't care about warning for you to update if you have installed a developer version
	if isDevVersion(curVer) {
		return ""
	}

	latestVer, err := getLatestVersionInfo()
	if err != nil {
		log.Errorf("error fetching latest version information: %v", err)
	}

	if latestVer.GT(curVer) {
		return pretty.LightYellow("warning: ") + getUpgradeMessage(latestVer, curVer)
	}

	return ""
}

func isDevVersion(s semver.Version) bool {
	if len(s.Build) != 0 {
		return true
	}

	if len(s.Pre) == 0 {
		return false
	}

	devStrings := regexp.MustCompile(`alpha|beta|dev|rc`)
	return !s.Pre[0].IsNum && devStrings.MatchString(s.Pre[0].VersionStr)
}

// getLatestVersionInfo returns information about the latest version of the CLI.
// It caches data from the server for a day.
func getLatestVersionInfo() (semver.Version, error) {
	cached, err := getCachedVersionInfo()
	if err == nil {
		return cached, nil
	}

	latestTag, err := gitutil.GetLatestTag()
	if err != nil {
		return semver.Version{}, err
	}

	latest, err := semver.ParseTolerant(latestTag)
	if err != nil {
		return semver.Version{}, err
	}

	if err = cacheVersionInfo(latest); err != nil {
		log.Errorf("failed to cache version info: %s", err)
	}

	return latest, nil
}

// getUpgradeMessage gets a message to display to a user instructing that
// they are out of date and how to move from current to latest.
func getUpgradeMessage(latest semver.Version, current semver.Version) string {
	cmd := getUpgradeCommand()

	msg := fmt.Sprintf("A new version of Kusion is available. To upgrade from version '%s' to '%s', ", current, latest)
	if cmd != "" {
		msg += "run \n   " + cmd + "\nor "
	}

	msg += "visit https://kusionstack.io/docs/user_docs/getting-started/install/ for manual instructions."
	msg += "\nNOTE: set env `KUSION_SKIP_UPDATE_CHECK` to `true` to skip version upgrade check."
	return msg
}

// getCachedVersionInfo reads cached information about the newest CLI version, returning the newest version available.
func getCachedVersionInfo() (semver.Version, error) {
	updateCheckFile, err := clipath.CachePath(".cached_version")
	if err != nil {
		return semver.Version{}, err
	}

	ts, err := times.Stat(updateCheckFile)
	if err != nil {
		return semver.Version{}, err
	}

	if time.Now().After(ts.ModTime().Add(24 * time.Hour)) {
		return semver.Version{}, errors.New("cached expired")
	}

	cached, err := os.ReadFile(updateCheckFile)
	if err != nil {
		return semver.Version{}, err
	}

	latest, err := semver.ParseTolerant(string(cached))
	if err != nil {
		return semver.Version{}, err
	}

	return latest, err
}

// cacheVersionInfo saves version information in a cache file to be looked up later.
func cacheVersionInfo(latest semver.Version) error {
	updateCheckFile, err := clipath.CachePath(".cached_version")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(updateCheckFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(latest.String())
	return err
}

// getUpgradeCommand returns a command that will upgrade the CLI to the newest version.
// If we can not determine how the CLI was installed, the empty string is returned.
func getUpgradeCommand() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	isKusionup, err := isKusionUpInstall(exe)
	if err != nil {
		log.Errorf("error determining if the running executable was installed with kusionup: %s", err)
	}
	if isKusionup {
		return "$ kusionup install"
	}
	return ""
}

// isKusionUpInstall returns true if the current running executable is running on linux based and was installed with kusionup.
// todo: delete embedding with kusionup
func isKusionUpInstall(exe string) (bool, error) {
	exePath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return false, err
	}

	curUser, err := user.Current()
	if err != nil {
		return false, err
	}

	prefix := filepath.Join(curUser.HomeDir, ".kusionup")
	return strings.HasPrefix(exePath, prefix), nil
}
