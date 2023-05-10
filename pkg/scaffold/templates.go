package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/spf13/afero"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/kfile"
)

// These are variables instead of constants in order that they can be set using the `-X`
// `ldflag` at build time, if necessary.
var (
	// The Git URL for Kusion program templates
	KusionTemplateGitRepository = "https://github.com/KusionStack/kusion-templates"
	// The branch name for the template repository
	kusionTemplateBranch = "main"
)

// TemplateRepository represents a repository of templates.
type TemplateRepository struct {
	Root         string // The full path to the root directory of the repository.
	SubDirectory string // The full path to the subdirectory within the repository.
	ShouldDelete bool   // Whether the root directory should be deleted.
}

// Delete deletes the template repository.
func (repo TemplateRepository) Delete() error {
	if repo.ShouldDelete {
		return os.RemoveAll(repo.Root)
	}
	return nil
}

// Templates lists the templates in the repository.
func (repo TemplateRepository) Templates() ([]Template, error) {
	path := repo.SubDirectory

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// If it's a file, look in its directory.
	if !info.IsDir() {
		path = filepath.Dir(path)
	}

	// See if there's a kusion.yaml in the directory.
	t, err := LoadTemplate(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if err == nil {
		return []Template{t}, nil
	}

	// Otherwise, read all subdirectories to find the ones
	// that contain a kusion.yaml.
	infos, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []Template
	for _, info := range infos {
		if info.IsDir() {
			name := info.Name()

			// Ignore the .git directory.
			if name == GitDir {
				continue
			}

			loadTemplate, err := LoadTemplate(filepath.Join(path, name))
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			} else if err == nil {
				result = append(result, loadTemplate)
			}
		}
	}
	return result, nil
}

// LoadTemplate returns a template from a path.
func LoadTemplate(path string) (Template, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Template{}, err
	}
	if !info.IsDir() {
		return Template{}, errors.Errorf("%s is not a directory", path)
	}

	proj, err := LoadProjectTemplate(filepath.Join(path, KusionYaml))
	if err != nil {
		return Template{}, err
	}

	t := Template{
		Dir:  path,
		Name: filepath.Base(path),
	}
	if proj != nil {
		t.ProjectTemplate = proj
	}

	return t, nil
}

// Template represents a project template.
type Template struct {
	// The directory containing kusion.yaml.
	Dir string `json:"dir,omitempty" yaml:"dir,omitempty"`
	// The name of the template.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	*ProjectTemplate
}

// RetrieveTemplates retrieves a "template repository" based on the specified name, path, or URL.
func RetrieveTemplates(templateNamePathOrURL string, online bool) (TemplateRepository, error) {
	if IsTemplateURL(templateNamePathOrURL) {
		return retrieveURLTemplates(templateNamePathOrURL, online)
	}
	if isTemplateFileOrDirectory(templateNamePathOrURL) {
		return retrieveFileTemplates(templateNamePathOrURL)
	}
	return retrieveKusionTemplates(templateNamePathOrURL, online)
}

// IsTemplateURL returns true if templateNamePathOrURL starts with "https://".
func IsTemplateURL(templateNamePathOrURL string) bool {
	return strings.HasPrefix(templateNamePathOrURL, "https://")
}

// retrieveURLTemplates retrieves the "template repository" at the specified URL.
func retrieveURLTemplates(rawurl string, online bool) (TemplateRepository, error) {
	if !online {
		return TemplateRepository{}, errors.Errorf("cannot use %s offline", rawurl)
	}

	var err error

	// Create a temp dir.
	var temp string
	if temp, err = os.MkdirTemp("", "kusion-template-"); err != nil {
		return TemplateRepository{}, err
	}

	var fullPath string
	if fullPath, err = workspace.RetrieveGitFolder(rawurl, temp); err != nil {
		return TemplateRepository{}, fmt.Errorf("failed to retrieve git folder: %w", err)
	}

	return TemplateRepository{
		Root:         temp,
		SubDirectory: fullPath,
		ShouldDelete: true,
	}, nil
}

// isTemplateFileOrDirectory returns true if templateNamePathOrURL is the name of a valid file or directory.
func isTemplateFileOrDirectory(templateNamePathOrURL string) bool {
	_, err := os.Stat(templateNamePathOrURL)
	return err == nil
}

// retrieveFileTemplates points to the "template repository" at the specified location in the file system.
func retrieveFileTemplates(path string) (TemplateRepository, error) {
	return TemplateRepository{
		Root:         path,
		SubDirectory: path,
		ShouldDelete: false,
	}, nil
}

// retrieveKusionTemplates retrieves the "template repository" for Kusion templates.
// Instead of retrieving to a temporary directory, the Kusion templates are managed from
// ~/.kusionup/current/templates.
func retrieveKusionTemplates(templateName string, online bool) (TemplateRepository, error) {
	templateName = strings.ToLower(templateName)

	// Cleanup the template directory.
	if err := cleanupLegacyTemplateDir(); err != nil {
		return TemplateRepository{}, err
	}

	// Get the template directory.
	templateDir, err := GetTemplateDir(ExternalTemplateDir)
	if err != nil {
		return TemplateRepository{}, err
	}

	// Ensure the template directory exists.
	if err := os.MkdirAll(templateDir, DefaultDirectoryPermission); err != nil {
		return TemplateRepository{}, err
	}

	if online {
		// Clone or update the kusion/templates repo.
		repo := KusionTemplateGitRepository
		branch := plumbing.NewBranchReferenceName(kusionTemplateBranch)
		err := gitutil.GitCloneOrPull(repo, branch, templateDir, false /*shallow*/)
		if err != nil {
			return TemplateRepository{}, fmt.Errorf("cloning templates repo: %w", err)
		}
	}

	subDir := templateDir
	if templateName != "" {
		subDir = filepath.Join(subDir, templateName)

		// Provide a nicer error message when the template can't be found (dir doesn't exist).
		_, err := os.Stat(subDir)
		if err != nil {
			if os.IsNotExist(err) {
				return TemplateRepository{}, newTemplateNotFoundError(templateDir, templateName)
			}
			log.Warnf("Explicitly ignoring and discarding error: %v", err)
		}
	}

	return TemplateRepository{
		Root:         templateDir,
		SubDirectory: subDir,
		ShouldDelete: false,
	}, nil
}

// cleanupLegacyTemplateDir deletes an existing ~/.kusionup/current/templates directory if it isn't a git repository.
func cleanupLegacyTemplateDir() error {
	templateDir, err := GetTemplateDir(ExternalTemplateDir)
	if err != nil {
		return err
	}

	// See if the template directory is a Git repository.
	repo, err := git.PlainOpen(templateDir)
	if err != nil {
		// If the repository doesn't exist, it's a legacy directory.
		// Delete the entire template directory and all children.
		if err == git.ErrRepositoryNotExists {
			return os.RemoveAll(templateDir)
		}

		return err
	}

	// The template directory is a Git repository. We want to make sure that it has the same remote as the one that
	// we want to pull from. If it doesn't have the same remote, we'll delete it, so that the clone later succeeds.
	url := KusionTemplateGitRepository

	remotes, err := repo.Remotes()
	if err != nil {
		return fmt.Errorf("getting template repo remotes: %w", err)
	}

	// If the repo exists, and it doesn't have exactly one remote that matches our URL, wipe the templates' directory.
	if len(remotes) != 1 || remotes[0] == nil || !strings.Contains(remotes[0].String(), url) {
		return os.RemoveAll(templateDir)
	}

	return nil
}

// GetTemplateDir returns the directory in which templates on the current machine are stored.
func GetTemplateDir(subDir string) (string, error) {
	kusionDir, err := kfile.KusionDataFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(kusionDir, subDir), nil
}

// newExistingFilesError returns a new error from a list of existing file names
// that would be overwritten.
func newExistingFilesError(existing []string) error {
	if len(existing) == 0 {
		return errors.New("no existing files")
	}

	message := "creating this template will make changes to existing files:\n"

	for _, file := range existing {
		message += fmt.Sprintf("  overwrite   %s\n", file)
	}

	message += "\nrerun the command and pass --force to accept and create"

	return errors.New(message)
}

// newTemplateNotFoundError returns an error for when the template doesn't exist,
// offering distance-based suggestions in the error message.
func newTemplateNotFoundError(templateDir string, templateName string) error {
	message := fmt.Sprintf("template '%s' not found", templateName)

	// Attempt to read the directory to offer suggestions.
	infos, err := os.ReadDir(templateDir)
	if err != nil {
		log.Errorf("os.ReadDir(%s) error: %v", templateDir, err)
		return errors.New(message)
	}

	// Get suggestions based on levenshtein distance.
	suggestions := []string{}
	const minDistance = 2
	op := levenshtein.DefaultOptions
	for _, info := range infos {
		distance := levenshtein.DistanceForStrings([]rune(templateName), []rune(info.Name()), op)
		if distance <= minDistance {
			suggestions = append(suggestions, info.Name())
		}
	}

	// Build-up error message with suggestions.
	if len(suggestions) > 0 {
		message += "\n\nDid you mean this?\n"
		for _, suggestion := range suggestions {
			message += fmt.Sprintf("\t%s\n", suggestion)
		}
	}

	return errors.New(message)
}

// Naming rules are backend-specific. However, we provide baseline sanitization for project names
// in this file. Though the backend may enforce stronger restrictions for a project name or description
// further down the line.
var (
	validProjectNameRegexp = regexp.MustCompile("^[A-Za-z0-9_.-]{1,100}$")
)

// ValidateProjectName ensures a project name is valid, if it is not it returns an error with a message suitable
// for display to an end user.
func ValidateProjectName(s string) error {
	if s == "" {
		return errors.New("A project name may not be empty")
	}

	if len(s) > 100 {
		return errors.New("A project name must be 100 characters or less")
	}

	if !validProjectNameRegexp.MatchString(s) {
		return errors.New("A project name may only contain alphanumeric, hyphens, underscores, and periods")
	}

	return nil
}

// TemplateConfig contains all config items to render the chosen project
type TemplateConfig struct {
	// ProjectName is project name, as well as root dir name
	ProjectName string `json:"projectName"`
	// ProjectConfig contains configuration in project level
	ProjectConfig map[string]interface{} `json:"projectConfig,omitempty"`
	// StacksConfig contains configuration in stack level, can be multi-stack or single-stack
	StacksConfig map[string]map[string]interface{} `json:"stacksConfig,omitempty"`
}

var (
	// file flag, create or upate
	CreateOrUpdate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	// default directory permission, 700
	DefaultDirectoryPermission os.FileMode = 0o700
	// default file permission, 600
	DefaultFilePermission os.FileMode = 0o600
)

// RenderLocalTemplate does the actual copy operation from source directory to a destination directory.
func RenderLocalTemplate(sourceDir, destDir string, force bool, tc *TemplateConfig) error {
	// Source FS
	srcFS := afero.NewMemMapFs()
	if err := ReadTemplate(sourceDir, srcFS); err != nil {
		return err
	}
	// Destination FS
	destFS := afero.NewMemMapFs()
	if err := RenderFSTemplate(srcFS, sourceDir, destFS, destDir, tc); err != nil {
		return err
	}
	// Write into disk
	return WriteToDisk(destFS, destDir, force)
}

// Read files' content from local dir into file system
func ReadTemplate(dir string, fs afero.Fs) error {
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		path := filepath.Join(dir, fileInfo.Name())
		if fileInfo.IsDir() {
			if err = fs.MkdirAll(path, DefaultDirectoryPermission); err != nil {
				return err
			}
			if err = ReadTemplate(path, fs); err != nil {
				return err
			}
		} else {
			// Read source file content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// Create or Update
			writer, err := fs.OpenFile(path, CreateOrUpdate, DefaultFilePermission)
			if err != nil {
				return err
			}
			defer func() {
				if closeErr := writer.Close(); err == nil && closeErr != nil {
					err = closeErr
				}
			}()
			// Write into FS
			if _, err := writer.Write(content); err != nil {
				return err
			}
		}
	}

	return nil
}

// RenderFSTemplate does the actual copy operation from source FS to destination FS.
func RenderFSTemplate(srcFS afero.Fs, srcDir string, destFS afero.Fs, destDir string, tc *TemplateConfig) error {
	// Read all sub dirs and files under srcDir
	fileInfos, err := afero.ReadDir(srcFS, srcDir)
	if err != nil {
		return err
	}
	for _, d := range fileInfos {
		src := filepath.Join(srcDir, d.Name())
		dest := filepath.Join(destDir, d.Name())
		if d.IsDir() {
			// Base dir or stack dir
			fileInfo, err := srcFS.Stat(filepath.Join(src, projectstack.StackFile))
			if err == nil && fileInfo.Mode().IsRegular() {
				// Project config can be overridden
				configs := make(map[string]interface{}, len(tc.ProjectConfig))
				for k, v := range tc.ProjectConfig {
					configs[k] = v
				}
				// Skip if stackConfigs are not provided
				stackConfigs, exits := tc.StacksConfig[d.Name()]
				if !exits {
					continue
				}
				// Merge and override project config
				for k, v := range stackConfigs {
					configs[k] = v
				}
				// Walk stack dir with merged configs
				err = walkFiles(srcFS, src, destFS, dest, configs)
				if err != nil {
					return err
				}
			} else {
				// Stack dir nested in 3rd level or even deeper, eg: meta_app/deployed_unit/stack_dir
				err = RenderFSTemplate(srcFS, src, destFS, dest, tc)
				if err != nil {
					return err
				}
			}
		} else {
			// project files. eg: project.yaml, README.md
			err = doFile(srcFS, src, destFS, dest, d.Name(), tc.ProjectConfig)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// walkFiles is a helper that walks the directories/files in a source directory
// and performs render for each item.
func walkFiles(srcFS afero.Fs, srcDir string, destFS afero.Fs, destDir string, config map[string]interface{}) error {
	// Create dest directory first
	err := destFS.MkdirAll(destDir, DefaultDirectoryPermission)
	if err != nil {
		return err
	}

	// Read files and dirs under srcDir
	dirs, err := afero.ReadDir(srcFS, srcDir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		src := filepath.Join(srcDir, d.Name())
		dest := filepath.Join(destDir, d.Name())
		if d.IsDir() {
			// Ignore the .git directory
			if d.Name() == GitDir {
				continue
			}
			// Walk sub dir, eg: template/prod/ci
			err = walkFiles(srcFS, src, destFS, dest, config)
			if err != nil {
				return err
			}
		} else {
			// render files, eg: base.k
			err = doFile(srcFS, src, destFS, dest, d.Name(), config)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// doFile render template file content and save it in destination file system
func doFile(srcFS afero.Fs, srcPath string, destFS afero.Fs, destPath, fileName string, config map[string]interface{}) error {
	// Read template
	srcContent, err := afero.ReadFile(srcFS, srcPath)
	if err != nil {
		return err
	}

	// Skip kusion.yaml
	if fileName == KusionYaml {
		return nil
	}

	// Render template file
	destContent, err := render(fileName, string(srcContent), config)
	if err != nil {
		return err
	}

	// Create or truncate the file
	writer, err := destFS.OpenFile(destPath, CreateOrUpdate, DefaultFilePermission)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := writer.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	// Write into destFS
	if _, err := writer.Write(destContent); err != nil {
		return err
	}
	return nil
}

// render parse content(string) with configMap(map[string]string) with go tmpl
func render(name string, content string, configMap map[string]interface{}) ([]byte, error) {
	temp := template.New(name)

	if _, err := temp.Parse(content); err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	if err := temp.Execute(out, configMap); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Walk destination file system and persist each file to local disk
func WriteToDisk(destFS afero.Fs, root string, force bool) error {
	return afero.Walk(destFS, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return mkdirWithForce(path, force)
		} else {
			content, err := afero.ReadFile(destFS, path)
			if err != nil {
				return err
			}
			err = writeAllBytes(path, content, force, DefaultFilePermission)
			if err != nil {
				// An existing file has shown up in between the dry run and the actual copy operation.
				if os.IsExist(err) {
					return newExistingFilesError([]string{filepath.Base(path)})
				}
				return err
			}
			return nil
		}
	})
}

// mkdirWithForce will ignore dir exists error when force is true
func mkdirWithForce(path string, force bool) error {
	if force {
		return os.MkdirAll(path, 0o700)
	}
	return os.Mkdir(path, 0o700)
}

// writeAllBytes writes the bytes to the specified file, with an option to overwrite.
func writeAllBytes(filename string, bytes []byte, overwrite bool, mode os.FileMode) error {
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}

	f, err := os.OpenFile(filename, flag, mode)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warnf("Explicitly ignoring and discarding error: %v", err)
		}
	}()
	_, err = f.Write(bytes)
	return err
}
