package engine

import (
	"fmt"

	yamlv2 "gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"

	kcl "kusionstack.io/kclvm-go"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/log"
	jsonUtil "kusionstack.io/kusion/pkg/util/json"
)

const MaxLogLength = 3751

func ResourcesYAML2Spec(resourcesYAML []kcl.KCLResult) (*models.Spec, error) {
	resources := []models.Resource{}

	for _, resourcesYamlMap := range resourcesYAML {
		// Convert kcl result to yaml string
		msg := jsonUtil.MustMarshal2String(resourcesYamlMap)
		if len(msg) > MaxLogLength {
			msg = msg[0:MaxLogLength]
		}

		log.Infof("convertKCLResult2Resources resource:%v", msg)
		// Using yamlv2.Marshal and yamlv3.Unmarshal is a workaround for the error "did not find expected '-' indicator" in unmarshalling yaml
		yamlByte, err := yamlv2.Marshal(resourcesYamlMap)
		if err != nil {
			return nil, fmt.Errorf("yaml marshal failed. %v,%w", jsonUtil.MustMarshal2String(resourcesYamlMap), err)
		}

		// Parse yaml string as Resource
		item := &models.Resource{}
		err = yamlv3.Unmarshal(yamlByte, item)
		if err != nil {
			return nil, err
		}

		resources = append(resources, *item)
	}

	return &models.Spec{Resources: resources}, nil
}
