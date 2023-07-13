package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/djherbis/times"
	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/apply"
	"kusionstack.io/kusion/pkg/cmd/check"
	"kusionstack.io/kusion/pkg/cmd/compile"
	"kusionstack.io/kusion/pkg/cmd/deps"
	"kusionstack.io/kusion/pkg/cmd/destroy"
	"kusionstack.io/kusion/pkg/cmd/env"
	cmdinit "kusionstack.io/kusion/pkg/cmd/init"
	"kusionstack.io/kusion/pkg/cmd/ls"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/cmd/version"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/gitutil"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/kfile"
	"kusionstack.io/kusion/pkg/util/pretty"
	versionInfo "kusionstack.io/kusion/pkg/version"
)

// NewDefaultKusionctlCommand creates the `kusionctl` command with default arguments
func NewDefaultKusionctlCommand() *cobra.Command {
	return NewDefaultKusionctlCommandWithArgs(os.Args, os.Stdin, os.Stdout, os.Stderr)
}

// NewDefaultKusionctlCommandWithArgs creates the `kusionctl` command with arguments
func NewDefaultKusionctlCommandWithArgs(args []string, in io.Reader, out, errOut io.Writer) *cobra.Command {
	kusionctl := NewKusionctlCmd(in, out, errOut)
	if len(args) <= 1 {
		return kusionctl
	}
	cmdPathPieces := args[1:]
	if _, _, err := kusionctl.Find(cmdPathPieces); err == nil {
		// sub command exist
		return kusionctl
	}
	return kusionctl
}

func NewKusionctlCmd(in io.Reader, out, err io.Writer) *cobra.Command {
	// Sending in 'nil' for the getLanguageFn() results in using LANGUAGE, LC_ALL,
	// LC_MESSAGES, or LANG environment variable in sequence.
	_ = i18n.LoadTranslations(i18n.DomainKusion, nil)

	updateCheckResult := make(chan string)

	var (
		rootShort = i18n.T(`Kusion manages the Kubernetes cluster by code`)

		rootLong = i18n.T(`
		Kusion is a cloud-native programmable technology stack, which manages the Kubernetes cluster by code.`)
	)

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:           "kusion",
		Short:         rootShort,
		Long:          templates.LongDesc(rootLong),
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// If we fail before we start the async update check, go ahead and close the
			// channel since we know it will never receive a value.
			var waitForUpdateCheck bool
			defer func() {
				if !waitForUpdateCheck {
					close(updateCheckResult)
				}
			}()

			// todo: delete env KUSION_SKIP_UPDATE_CHECK, only show need updating info when run kusion version
			if v := os.Getenv("KUSION_SKIP_UPDATE_CHECK"); v == "true" {
				log.Infof("skipping update check")
			} else {
				// Run the version check in parallel so that it doesn't block executing the command.
				// If there is a new version to report, we will do so after the command has finished.
				waitForUpdateCheck = true
				go func() {
					updateCheckResult <- checkForUpdate()
					close(updateCheckResult)
				}()
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			checkVersionMsg, ok := <-updateCheckResult
			if ok && checkVersionMsg != "" {
				fmt.Println(checkVersionMsg)
			}
		},
	}

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	groups := templates.CommandGroups{
		{
			Message: "Configuration Commands:",
			Commands: []*cobra.Command{
				cmdinit.NewCmdInit(),
				compile.NewCmdCompile(),
				check.NewCmdCheck(),
				ls.NewCmdLs(),
				deps.NewCmdDeps(),
			},
		},
		{
			Message: "RuntimeMap Commands:",
			Commands: []*cobra.Command{
				preview.NewCmdPreview(),
				apply.NewCmdApply(),
				destroy.NewCmdDestroy(),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	templates.ActsAsRootCommand(cmds, filters, groups...)
	// Add other subcommands
	// TODO: add plugin subcommand
	// cmds.AddCommand(plugin.NewCmdPlugin(f, ioStreams))
	cmds.AddCommand(version.NewCmdVersion())
	cmds.AddCommand(env.NewCmdEnv())

	return cmds
}

// checkForUpdate checks to see if the CLI needs to be updated,
// and if so emits a warning, as well as information as to how it can be upgraded.
func checkForUpdate() string {
	curVer, err := semver.ParseTolerant(versionInfo.ReleaseVersion())
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

// getCachedVersionInfo reads cached information about the newest CLI version, returning the newest version available.
func getCachedVersionInfo() (semver.Version, error) {
	updateCheckFile, err := kfile.GetCachedVersionFilePath()
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
	updateCheckFile, err := kfile.GetCachedVersionFilePath()
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
