package svc

import (
	"errors"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/scaffold"
)

// Template contains complete information of a template
type Template struct {
	// Name of the template.
	Name string `json:"name" yaml:"name"`
	// Dir is the directory containing kusion.yaml.
	Dir string `json:"dir" yaml:"dir"`

	scaffold.ProjectTemplate
}

var (
	ErrEmptyTemplateRepoURL = errors.New("empty template repo url")
	ErrEmptyTemplatePath    = errors.New("empty template path")
)

// ListTemplateOptions is the options for querying templates.
type ListTemplateOptions struct {
	// Online indicates querying templates from online repo or local path.
	Online bool `json:"online,omitempty" yaml:"online,omitempty"`
	// URL is the online repo url, works when Online is true.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// Path is the local path to find the templates, works when Online is false.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// Validate is used to check the validation of ListTemplateOptions.
func (o *ListTemplateOptions) Validate() error {
	if o.Online && o.URL == "" {
		return ErrEmptyTemplateRepoURL
	}
	if !o.Online && o.Path == "" {
		return ErrEmptyTemplatePath
	}
	return nil
}

// ListTemplates queries templates from online repo or local path.
func ListTemplates(o *ListTemplateOptions) ([]*Template, error) {
	// Validate listTemplateOptions
	if err := o.Validate(); err != nil {
		return nil, WrapInvalidArgumentErr(err)
	}

	// retrieve the template repo.
	var templateNamePathOrURL string
	if o.Online {
		templateNamePathOrURL = o.URL
	} else {
		templateNamePathOrURL = o.Path
	}
	repo, err := scaffold.RetrieveTemplates(templateNamePathOrURL, o.Online)
	if err != nil {
		return nil, WrapInternalErr(err)
	}
	defer func() {
		if err = repo.Delete(); err != nil {
			log.Warnf("explicitly ignoring and discarding error, %v", err)
		}
	}()

	// list the templates from the repo.
	ts, err := repo.Templates()
	if err != nil {
		return nil, WrapInternalErr(err)
	}
	templates := make([]*Template, len(ts))
	for i, t := range ts {
		templates[i] = &Template{
			Name:            t.Name,
			Dir:             t.Dir,
			ProjectTemplate: *t.ProjectTemplate,
		}
	}
	return templates, nil
}
