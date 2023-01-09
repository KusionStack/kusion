package tfops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/spf13/afero"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/log"
)

const (
	HCLMAINFILE = "main.tf.json"
	HCLLOCKFILE = ".terraform.hcl.lock"
	TFSTATEFILE = "terraform.tfstate"
)

type WorkSpace struct {
	resource   *models.Resource
	fs         afero.Afero
	stackDir   string
	tfCacheDir string
}

// SetResource set workspace resource
func (w *WorkSpace) SetResource(resource *models.Resource) {
	w.resource = resource
}

// SetFS set filesystem
func (w *WorkSpace) SetFS(fs afero.Afero) {
	w.fs = fs
}

// SetStackDir set workspace work directory.
func (w *WorkSpace) SetStackDir(stackDir string) {
	w.stackDir = stackDir
}

// SetCacheDir set tf cache work directory.
func (w *WorkSpace) SetCacheDir(cacheDir string) {
	w.tfCacheDir = cacheDir
}

func NewWorkSpace(fs afero.Afero) *WorkSpace {
	return &WorkSpace{
		fs: fs,
	}
}

// WriteHCL convert kusion Resource to HCL json
// and write hcl json to main.tf.json
func (w *WorkSpace) WriteHCL() error {
	provider := strings.Split(w.resource.Extensions["provider"].(string), "/")
	resourceType := w.resource.Extensions["resourceType"].(string)
	resourceNames := strings.Split(w.resource.ResourceKey(), ":")

	m := map[string]interface{}{
		"terraform": map[string]interface{}{
			"required_providers": map[string]interface{}{
				provider[len(provider)-2]: map[string]string{
					"source":  strings.Join(provider[:len(provider)-1], "/"),
					"version": provider[len(provider)-1],
				},
			},
		},
		"provider": map[string]interface{}{
			provider[len(provider)-2]: w.resource.Extensions["providerMeta"],
		},
		"resource": map[string]interface{}{
			resourceType: map[string]interface{}{
				resourceNames[len(resourceNames)-1]: w.resource.Attributes,
			},
		},
	}
	hclMain, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal hcl main error: %v", err)
	}

	_, err = w.fs.Stat(w.tfCacheDir)

	if err != nil {
		if os.IsNotExist(err) {
			if err := w.fs.MkdirAll(w.tfCacheDir, os.ModePerm); err != nil {
				return fmt.Errorf("create workspace error: %v", err)
			}
		} else {
			return err
		}
	}
	err = w.fs.WriteFile(filepath.Join(w.tfCacheDir, HCLMAINFILE), hclMain, 0o600)
	if err != nil {
		return fmt.Errorf("write hcl main.tf.json error: %v", err)
	}

	return nil
}

// WriteTFState writes TFState to the file, this function is for terraform apply refresh only
func (w *WorkSpace) WriteTFState(priorState *models.Resource) error {
	provider := strings.Split(priorState.Extensions["provider"].(string), "/")
	m := map[string]interface{}{
		"version": 4,
		"resources": []map[string]interface{}{
			{
				"mode":     "managed",
				"type":     priorState.Extensions["resourceType"].(string),
				"name":     priorState.ID,
				"provider": fmt.Sprintf("provider[\"%s\"]", strings.Join(provider[:len(provider)-1], "/")),
				"instances": []map[string]interface{}{
					{
						"attributes": priorState.Attributes,
					},
				},
			},
		},
	}
	hclState, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal hcl state error: %v", err)
	}

	err = w.fs.WriteFile(filepath.Join(w.tfCacheDir, TFSTATEFILE), hclState, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write hcl  error: %v", err)
	}
	return nil
}

// InitWorkSpace init terraform runtime workspace
func (w *WorkSpace) InitWorkSpace(ctx context.Context) error {
	chdir := fmt.Sprintf("-chdir=%s", w.tfCacheDir)
	cmd := exec.CommandContext(ctx, "terraform", chdir, "init")
	cmd.Dir = w.stackDir
	_, err := cmd.Output()
	if e, ok := err.(*exec.ExitError); ok {
		return errors.New(string(e.Stderr))
	}
	return nil
}

// Apply with the terraform cli apply command
func (w *WorkSpace) Apply(ctx context.Context) (*TFState, error) {
	chdir := fmt.Sprintf("-chdir=%s", w.tfCacheDir)
	err := w.CleanAndInitWorkspace(ctx, chdir)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "terraform", chdir, "apply", "-auto-approve", "-json", "-lock=false")
	cmd.Dir = w.stackDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, TFError(out)
	}
	s, err := w.RefreshOnly(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform read state error: %v", err)
	}
	return s, err
}

// Read make terraform show call. Return terraform state model
// TODO: terraform show livestate.
func (w *WorkSpace) Read(ctx context.Context) (*TFState, error) {
	_, err := w.fs.Stat(filepath.Join(w.tfCacheDir, "terraform.tfstate"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	chdir := fmt.Sprintf("-chdir=%s", w.tfCacheDir)
	cmd := exec.CommandContext(ctx, "terraform", chdir, "show", "-json")
	cmd.Dir = w.stackDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, TFError(out)
	}
	s := &TFState{}
	if err = json.Unmarshal(out, s); err != nil {
		return nil, fmt.Errorf("json umarshal state failed: %v", err)
	}
	return s, nil
}

// Refresh Sync Terraform State
func (w *WorkSpace) RefreshOnly(ctx context.Context) (*TFState, error) {
	chdir := fmt.Sprintf("-chdir=%s", w.tfCacheDir)
	err := w.CleanAndInitWorkspace(ctx, chdir)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "terraform", chdir, "apply", "-auto-approve", "-json", "--refresh-only", "-lock=false")
	cmd.Dir = w.stackDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, TFError(out)
	}
	s, err := w.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform read state error: %v", err)
	}
	return s, err
}

// Destroy make terraform destroy call.
func (w *WorkSpace) Destroy(ctx context.Context) error {
	chdir := fmt.Sprintf("-chdir=%s", w.tfCacheDir)
	cmd := exec.CommandContext(ctx, "terraform", chdir, "destroy", "-auto-approve")
	cmd.Dir = w.stackDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return TFError(out)
	}
	return nil
}

// GetProvider get provider addr from terraform lock file.
// return provider addr and errors
// eg. registry.terraform.io/hashicorp/local/2.2.3
func (w *WorkSpace) GetProvider() (string, error) {
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCLFile(filepath.Join(w.tfCacheDir, ".terraform.lock.hcl"))
	if diags != nil {
		return "", errors.New(diags.Error())
	}
	body := hclFile.Body
	content, diags := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "provider",
				LabelNames: []string{"source_addr"},
			},
		},
	})
	if diags != nil {
		return "", errors.New(diags.Error())
	}
	rawAddr := content.Blocks[0].Labels[0]

	block := content.Blocks[0]
	providerVersion, _ := block.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "version", Required: true},
			{Name: "constraints"},
			{Name: "hashes"},
		},
	})
	expr := providerVersion.Attributes["version"].Expr
	var rawVersion string
	diags = gohcl.DecodeExpression(expr, nil, &rawVersion)
	if diags != nil {
		return "", errors.New(diags.Error())
	}

	providerAddr := fmt.Sprintf("%s/%s", rawAddr, rawVersion)
	return providerAddr, nil
}

// CleanAndInitWorkspace will clean up the provider cache and reinitialize the workspace
// when the provider version or hash is updated.
func (w *WorkSpace) CleanAndInitWorkspace(ctx context.Context, chdir string) error {
	isHashUpdate := w.checkHashUpdate(ctx, chdir)
	isVersionUpdate, err := w.checkVersionUpdate(ctx)
	if err != nil {
		return fmt.Errorf("check provider version failed: %v", err)
	}

	// If the provider hash or version changes, delete the tf cache and reinitialize.
	if isHashUpdate || isVersionUpdate {
		log.Info("provider hash or version change.")
		os.Remove(filepath.Join(w.tfCacheDir, ".terraform.lock.hcl"))
		os.Remove(filepath.Join(w.tfCacheDir, ".terraform"))
		err := w.InitWorkSpace(ctx)
		if err != nil {
			return fmt.Errorf("init terraform workspace failed: %v", err)
		}
	}
	return nil
}

// checkHashUpdate checks whether the provider hash has changed, and returns true if changed
func (w *WorkSpace) checkHashUpdate(ctx context.Context, chdir string) bool {
	cmd := exec.CommandContext(ctx, "terraform", chdir, "providers", "lock")
	cmd.Dir = w.stackDir
	output, _ := cmd.Output()
	return strings.Contains(string(output), "Terraform has updated the lock file")
}

// checkVersionUpdate checks whether the provider version has changed, and returns true if changed
func (w *WorkSpace) checkVersionUpdate(ctx context.Context) (bool, error) {
	providerAddr, err := w.GetProvider()
	if err != nil {
		return false, fmt.Errorf("provider get version failed: %v", err)
	}
	return providerAddr != w.resource.Extensions["provider"].(string), nil
}
