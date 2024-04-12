package mod

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
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
		# Push a module of current OS arch to an OCI Registry using a token
		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --version=1.0.0 --creds <YOUR_TOKEN>

		# Push a module of specific OS arch to an OCI Registry using a token
		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --os-arch==darwin/arm64 --version=1.0.0 --creds <YOUR_TOKEN>
		
		# Push a module to an OCI Registry using a credentials in <YOUR_USERNAME>:<YOUR_TOKEN> format. 
		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --version=1.0.0 --creds <YOUR_USERNAME>:<YOUR_TOKEN>

		# Push a release candidate without marking it as the latest stable
		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --version=1.0.0-rc.1 --latest=false

		# Push a module with custom OCI annotations
		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --version=1.0.0 \
		  --annotation='org.opencontainers.image.documentation=https://app.org/docs'

		# Push and sign a module with Cosign (the cosign binary must be present in PATH)
		export COSIGN_PASSWORD=password
  		kusion mod push /path/to/my-module oci://ghcr.io/org/my-module --version=1.0.0 \
		  --sign=cosign --cosign-key=/path/to/cosign.key`)
)

// LatestVersion is the tag name that
// denotes the latest stable version of a module.
const LatestVersion = "latest"

// PushModFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// PushModOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type PushModFlags struct {
	Version          string
	Latest           bool
	OSArch           string
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
	OSArch     string
	Version    string
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions(args, flags.IOStreams)
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *PushModFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flags.Version, "version", "v", "", "The version of the module e.g. '1.0.0' or '1.0.0-rc.1'.")
	cmd.Flags().StringVar(&flags.OSArch, "os-arch", "", "The os arch of the module e.g. 'darwin/arm64', 'linux/amd64'.")
	cmd.Flags().BoolVar(&flags.Latest, "latest", flags.Latest, "Tags the current version as the latest stable module version.")
	cmd.Flags().StringVar(&flags.Credentials, "creds", flags.Credentials,
		"The credentials token for the OCI registry in <YOUR_TOKEN> or <YOUR_USERNAME>:<YOUR_TOKEN> format.")
	cmd.Flags().StringVar(&flags.Sign, "sign", flags.Sign, "Signs the module with the specified provider.")
	cmd.Flags().StringVar(&flags.CosignKey, "cosign-key", flags.CosignKey, "The Cosign private key for signing the module.")
	cmd.Flags().BoolVar(&flags.InsecureRegistry, "insecure-registry", flags.InsecureRegistry, "If true, allows connecting to a OCI registry without TLS or with self-signed certificates.")
	cmd.Flags().StringSliceVarP(&flags.Annotations, "annotations", "a", flags.Annotations,
		"Set custom OCI annotations in '<KEY>=<VALUE>' format.")
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *PushModFlags) ToOptions(args []string, ioStreams genericiooptions.IOStreams) (*PushModOptions, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("path to module and OCI registry url are required")
	}

	// Prepare metadata
	if flags.OSArch == "" {
		// set as the current OS arch
		flags.OSArch = runtime.GOOS + "/" + runtime.GOARCH
	}
	osArch := strings.Split(flags.OSArch, "/")

	version := flags.Version
	if _, err := semver.StrictNewVersion(version); err != nil {
		return nil, fmt.Errorf("version is not in semver format: %w", err)
	}
	fullURL := fmt.Sprintf("%s-%s_%s:%s", args[1], osArch[0], osArch[1], version)

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

	meta := metadata.Metadata{
		Created:     info.CommitDate,
		Source:      info.RemoteURL,
		Revision:    info.Commit,
		Annotations: annotations,
		Platform: &v1.Platform{
			OS:           osArch[0],
			Architecture: osArch[1],
		},
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
		ociclient.WithPlatform(meta.Platform),
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
	sp := &pretty.SpinnerT
	sp, _ = sp.Start("building the module binary...")
	defer func() {
		_ = sp.Stop()
	}()

	targetDir, err := o.buildModule()
	defer os.RemoveAll(targetDir)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Copy to temp module dir and push artifact to OCI repository
	err = ioutil.CopyDir(targetDir, o.ModulePath, func(path string) bool {
		skipDirs := []string{filepath.Join(o.ModulePath, ".git"), filepath.Join(o.ModulePath, "src")}

		// skip files in skipDirs
		for _, dir := range skipDirs {
			if strings.HasPrefix(path, dir) {
				return true
			}
		}
		return false
	})
	if err != nil {
		return err
	}

	sp.Info("pushing the module...")
	digest, err := o.Client.Push(ctx, o.OCIUrl, targetDir, o.Metadata, nil)
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

// build target os arch module binary
func (o *PushModOptions) buildModule() (string, error) {
	// First build executable binary via compilation
	// Create temp module dir for later tar operation
	targetDir, err := os.MkdirTemp("", filepath.Base(o.ModulePath))
	if err != nil {
		return "", err
	}

	moduleSrc := filepath.Join(o.ModulePath, "src")
	goFileSearchPattern := filepath.Join(moduleSrc, "*.go")

	// OCIUrl example: oci://ghcr.io/org/my-module-linux_amd64:0.1.0
	split := strings.Split(o.OCIUrl, "/")
	nameVersion := strings.Split(split[len(split)-1], ":")
	name := nameVersion[0]
	version := nameVersion[1]

	if matches, err := filepath.Glob(goFileSearchPattern); err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no go source code files found for 'go build' matching %s", goFileSearchPattern)
	}

	goBin, err := executable.FindExecutable("go")
	if err != nil {
		return "", fmt.Errorf("unable to find executable 'go' binary: %w", err)
	}

	// prepare platform
	if o.Metadata.Platform == nil {
		return "", fmt.Errorf("platform is not set in metadata")
	}
	pOS := o.Metadata.Platform.OS
	pArch := o.Metadata.Platform.Architecture
	output := filepath.Join(targetDir, "_dist", pOS, pArch, "kusion-module-"+name+"_"+version)
	if strings.Contains(o.OSArch, "windows") {
		output = filepath.Join(targetDir, "_dist", pOS, pArch, "kusion-module-"+name+"_"+version+".exe")
	}

	path, err := buildBinary(goBin, pOS, pArch, moduleSrc, output, o.IOStreams)
	if err != nil {
		return "", fmt.Errorf("failed to build the module %w", err)
	}

	return filepath.Dir(path), nil
}

// This function takes a file target to specify where to compile to.
// If `outfile` is "", the binary is compiled to a new temporary file.
// This function returns the path of the file that was produced.
func buildBinary(goBin, operatingSystem, arch, srcDirectory, outfile string, ioStreams genericiooptions.IOStreams) (string, error) {
	if outfile == "" {
		// If no outfile is supplied, write the Go binary to a temporary file.
		f, err := os.CreateTemp("", "generator.*")
		if err != nil {
			return "", fmt.Errorf("unable to create go program temp file: %w", err)
		}

		if err := f.Close(); err != nil {
			return "", fmt.Errorf("unable to close go program temp file: %w", err)
		}
		outfile = f.Name()
	}

	extraEnvs := []string{
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", operatingSystem),
		fmt.Sprintf("GOARCH=%s", arch),
	}

	buildCmd := exec.Command(goBin, "build", "-o", outfile)
	buildCmd.Dir = srcDirectory
	buildCmd.Env = append(os.Environ(), extraEnvs...)
	buildCmd.Stdout, buildCmd.Stderr = ioStreams.Out, ioStreams.ErrOut

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("unable to run `go build`: %w", err)
	}

	return outfile, nil
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
