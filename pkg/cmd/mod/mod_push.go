package mod

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/oci"
	ociclient "kusionstack.io/kusion/pkg/oci/client"
	"kusionstack.io/kusion/pkg/oci/metadata"
	"kusionstack.io/kusion/pkg/util/executable"
	"kusionstack.io/kusion/pkg/util/gitutil"
	"kusionstack.io/kusion/pkg/util/i18n"
	ioutil "kusionstack.io/kusion/pkg/util/io"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var (
	pushLong = i18n.T(`
		The push command packages the module as an OCI artifact and pushes it to the
		OCI registry using the version as the image tag.`)

	pushExample = i18n.T(`
		# Push a module to GitHub Container Registry using a GitHub token
		kusion mod push /path/to/my-module oci://ghcr.io/org/kusionstack/my-module --version=1.0.0 --creds $GITHUB_TOKEN

		# Push a release candidate without marking it as the latest stable
		kusion mod push /path/to/my-module oci://ghcr.io/kusionstack/my-module --version=1.0.0-rc.1 --latest=false

		# Push a module with custom OCI annotations
		kusion mod push /path/to/my-module oci://ghcr.io/org/kusionstack/my-module --version=1.0.0 \
		  --annotation='org.opencontainers.image.documentation=https://app.org/docs'

		# Push and sign a module with Cosign (the cosign binary must be present in PATH)
		export COSIGN_PASSWORD=password
  		kusion mod push /path/to/my-module oci://ghcr.io/org/kusionstack/my-module --version=1.0.0 \
		  --sign=cosign --cosign-key=/path/to/cosign.key`)
)

// LatestVersion is the tag name that
// denotes the latest stable version of a module.
const LatestVersion = "latest"

// All supported platforms, to reduce module package size, only support widely used os and arch.
var supportPlatforms = []string{
	"linux/amd64", "darwin/amd64", "windows/amd64", "darwin/arm64",
}

// PushModFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// PushModOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type PushModFlags struct {
	Version          string
	Latest           bool
	Annotations      []string
	Credentials      string
	Sign             string
	CosignKey        string
	InsecureRegistry bool

	genericiooptions.IOStreams
}

// PushModOptions is a set of options that allows you to push module. This is the object reflects the
// runtime needs of a `mod push` command, making the logic itself easy to unit test.
type PushModOptions struct {
	ModulePath string
	OCIUrl     string
	Latest     bool
	Sign       string
	CosignKey  string

	Client   *ociclient.Client
	Metadata metadata.Metadata

	genericiooptions.IOStreams
}

// NewPushModFlags returns a default PushModFlags.
func NewPushModFlags(ioStreams genericiooptions.IOStreams) *PushModFlags {
	return &PushModFlags{
		IOStreams:        ioStreams,
		Latest:           true,
		InsecureRegistry: false,
	}
}

// NewCmdPush returns an initialized Command instance for the 'mod push' sub command.
func NewCmdPush(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewPushModFlags(ioStreams)

	cmd := &cobra.Command{
		Use:                   "push [MODULE PATH] [OCI REPOSITORY URL]",
		DisableFlagsInUseLine: true,
		Short:                 "Push a module to OCI registry",
		Long:                  pushLong,
		Example:               pushExample,
		Run: func(cmd *cobra.Command, args []string) {
			o, err := flags.ToOptions(args, flags.IOStreams)
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *PushModFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flags.Version, "version", "v", flags.Version, "The version of the module e.g. '1.0.0' or '1.0.0-rc.1'.")
	cmd.Flags().BoolVar(&flags.Latest, "latest", flags.Latest, "Tags the current version as the latest stable module version.")
	cmd.Flags().StringVar(&flags.Credentials, "creds", flags.Credentials, "The credentials for the OCI registry in '<username>[:<password>]' format.")
	cmd.Flags().StringVar(&flags.Sign, "sign", flags.Sign, "Signs the module with the specified provider.")
	cmd.Flags().StringVar(&flags.CosignKey, "cosign-key", flags.CosignKey, "The Cosign private key for signing the module.")
	cmd.Flags().BoolVar(&flags.InsecureRegistry, "insecure-registry", flags.InsecureRegistry, "If true, allows connecting to a OCI registry without TLS or with self-signed certificates.")
	cmd.Flags().StringSliceVarP(&flags.Annotations, "annotations", "a", flags.Annotations, "Set custom OCI annotations in '<key>=<value>' format.")
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *PushModFlags) ToOptions(args []string, ioStreams genericiooptions.IOStreams) (*PushModOptions, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("path to module and OCI registry url are required")
	}

	version := flags.Version
	if _, err := semver.StrictNewVersion(version); err != nil {
		return nil, fmt.Errorf("version is not in semver format: %w", err)
	}
	fullURL := fmt.Sprintf("%s:%s", args[1], version)

	// If creds in <token> format, creds must be base64 encoded
	if len(flags.Credentials) != 0 && !strings.Contains(flags.Credentials, ":") {
		flags.Credentials = base64.StdEncoding.EncodeToString([]byte(flags.Credentials))
	}

	// Parse custom annotations
	annotations, err := metadata.ParseAnnotations(flags.Annotations)
	if err != nil {
		return nil, err
	}
	annotations[metadata.AnnotationVersion] = version

	// Detect git repository to get basic git information
	var info gitutil.Info
	repoRoot, err := detectGitRepository(args[0])
	if err == nil && len(repoRoot) != 0 {
		info = gitutil.Get(repoRoot)
	}

	// Prepare metadata
	meta := metadata.Metadata{
		Created:     info.CommitDate,
		Source:      info.RemoteURL,
		Revision:    info.Commit,
		Annotations: annotations,
	}
	if len(meta.Created) == 0 {
		ct := time.Now().UTC()
		meta.Created = ct.Format(time.RFC3339)
	}

	// Construct OCI repository client
	opts := []ociclient.ClientOption{
		ociclient.WithUserAgent(oci.UserAgent),
		ociclient.WithCredentials(flags.Credentials),
		ociclient.WithInsecure(flags.InsecureRegistry),
	}
	client := ociclient.NewClient(opts...)

	opt := &PushModOptions{
		ModulePath: args[0],
		OCIUrl:     fullURL,
		Latest:     flags.Latest,
		Sign:       flags.Sign,
		CosignKey:  flags.CosignKey,
		Client:     client,
		Metadata:   meta,
		IOStreams:  ioStreams,
	}

	return opt, nil
}

// Validate verifies if PushModOptions are valid and without conflicts.
func (o *PushModOptions) Validate() error {
	if fileInfo, err := os.Stat(o.ModulePath); err != nil || !fileInfo.IsDir() {
		return fmt.Errorf("no module found at path %s", o.ModulePath)
	}

	// TODO: add oci url validation
	return nil
}

// Run executes the `mod push` command.
func (o *PushModOptions) Run() error {
	// First build executable binary via compilation
	// Create temp module dir for later tar operation
	tempModuleDir, err := os.MkdirTemp("", filepath.Base(o.ModulePath))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempModuleDir)

	sp := &pretty.SpinnerT
	sp, _ = sp.Start("building the module binary...")
	defer func() {
		_ = sp.Stop()
	}()

	generatorSourceDir := filepath.Join(o.ModulePath, "src")
	err = buildGeneratorCrossPlatforms(generatorSourceDir, tempModuleDir, o.IOStreams)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Copy to temp module dir and push artifact to OCI repository
	err = ioutil.CopyDir(tempModuleDir, o.ModulePath, func(path string) bool {
		return strings.Contains(path, "src")
	})
	if err != nil {
		return err
	}

	sp.Info("pushing the module...")
	digest, err := o.Client.Push(ctx, o.OCIUrl, tempModuleDir, o.Metadata, nil)
	if err != nil {
		return err
	}

	// Tag latest version if required
	if o.Latest {
		if err = o.Client.Tag(ctx, digest, LatestVersion); err != nil {
			return fmt.Errorf("tagging module version as latest failed: %w", err)
		}
	}
	sp.Info("pushed successfully\n")
	_ = sp.Stop()

	// Signs the module with specific provider
	if len(o.Sign) != 0 {
		err = oci.SignArtifact(o.Sign, digest, o.CosignKey)
		if err != nil {
			return err
		}
	}

	return nil
}

// This function loops through all support platforms to build target binary.
func buildGeneratorCrossPlatforms(generatorSrcDir, targetDir string, ioStreams genericiooptions.IOStreams) error {
	goFileSearchPattern := filepath.Join(generatorSrcDir, "*.go")
	if matches, err := filepath.Glob(goFileSearchPattern); err != nil || len(matches) == 0 {
		return fmt.Errorf("no go source code files found for 'go build' matching %s", goFileSearchPattern)
	}

	gobin, err := executable.FindExecutable("go")
	if err != nil {
		return fmt.Errorf("unable to find 'go' executable: %w", err)
	}

	var wg sync.WaitGroup
	var failMu sync.Mutex
	failed := false

	// Build in parallel to reduce module push time
	for _, platform := range supportPlatforms {
		wg.Add(1)
		go func(plat string) {
			partialPath := strings.Replace(plat, "/", "-", 1)
			output := filepath.Join(targetDir, "_dist", partialPath, "generator")
			if strings.Contains(plat, "windows") {
				output = filepath.Join(targetDir, "_dist", partialPath, "generator.exe")
			}
			f := false
			buildErr := buildGenerator(gobin, plat, generatorSrcDir, output, ioStreams)
			if buildErr != nil {
				fmt.Printf("failed to build with %s\n", plat)
				f = true
			}
			failMu.Lock()
			failed = failed || f
			failMu.Unlock()
			wg.Done()
		}(platform)
	}
	wg.Wait()

	if failed {
		return fmt.Errorf("failed to build generator bin")
	}

	return nil
}

// This function takes a file target to specify where to compile to.
// If `outfile` is "", the binary is compiled to a new temporary file.
// This function returns the path of the file that was produced.
func buildGenerator(gobin, platform, generatorDirectory, outfile string, ioStreams genericiooptions.IOStreams) error {
	if outfile == "" {
		// If no outfile is supplied, write the Go binary to a temporary file.
		f, err := os.CreateTemp("", "generator.*")
		if err != nil {
			return fmt.Errorf("unable to create go program temp file: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("unable to close go program temp file: %w", err)
		}
		outfile = f.Name()
	}

	osArch := strings.Split(platform, "/")
	extraEnvs := []string{
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", osArch[0]),
		fmt.Sprintf("GOARCH=%s", osArch[1]),
	}

	buildCmd := exec.Command(gobin, "build", "-o", outfile)
	buildCmd.Dir = generatorDirectory
	buildCmd.Env = append(os.Environ(), extraEnvs...)
	buildCmd.Stdout, buildCmd.Stderr = ioStreams.Out, ioStreams.ErrOut

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("unable to run `go build`: %w", err)
	}

	return nil
}

// detectGitRepository detects existence of .git with target path.
func detectGitRepository(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for {
		st, err := os.Stat(filepath.Join(path, ".git"))
		if err == nil && st.IsDir() {
			break
		}
		old := path
		path = filepath.Dir(path)
		if old == path {
			return "", fmt.Errorf("could not detect git repository root")
		}
	}
	return path, nil
}
