package utils

import (
	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	yamlv3 "gopkg.in/yaml.v3"
)

// LoadFile processes the provided input location to load it as one of the
// supported document formats, or plain text if nothing else works.
func LoadFile(yaml, location string) (ytbx.InputFile, error) {
	var (
		documents []*yamlv3.Node
		data      []byte
		err       error
	)

	data = []byte(yaml)
	if documents, err = ytbx.LoadDocuments(data); err != nil {
		return ytbx.InputFile{}, wrap.Errorf(err, "unable to parse data %v", data)
	}

	return ytbx.InputFile{
		Location:  location,
		Documents: documents,
	}, nil
}
