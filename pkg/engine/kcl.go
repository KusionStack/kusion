package engine

import (
	"fmt"

	"github.com/pkg/errors"
	"kusionstack.io/KCLVM/kclvm-go/api/kcl"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	jsonUtil "kusionstack.io/kusion/pkg/util/json"
)

const MaxLogLength = 3751

func ConvertKCLResult2Resources(resourceYAMLs []kcl.KCLResult) (*manifest.Manifest, error) {
	var resources []states.ResourceState

	for _, resourcesYamlMap := range resourceYAMLs {
		// Convert kcl result to yaml string
		msg := jsonUtil.MustMarshal2String(resourcesYamlMap)
		if len(msg) > MaxLogLength {
			msg = msg[0:MaxLogLength]
		}
		log.Infof("convertKCLResult2Resources resource:%v", msg)
		yamlByte, err := yaml.Marshal(resourcesYamlMap)
		if err != nil {
			return nil, fmt.Errorf("yaml marshal failed. %v,%w", jsonUtil.MustMarshal2String(resourcesYamlMap), err)
		}

		// Parse yaml string as Resource
		item := &states.ResourceState{}
		err = yaml.Unmarshal(yamlByte, item)
		if err != nil {
			return nil, err
		}
		// TODO: any other better judgement here?
		if item.Attributes == nil {
			item, _, err = NewRequestResourceForKubernetes(resourcesYamlMap)
			if err != nil {
				return nil, errors.Wrap(err, "compile result format error (neither kubernetes nor engine resource format)")
			}
		}

		resources = append(resources, *item)
	}

	return &manifest.Manifest{resources}, nil
}
