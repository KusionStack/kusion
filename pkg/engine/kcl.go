package engine

import (
	"encoding/json"

	kcl "kcl-lang.io/kcl-go"

	"kusionstack.io/kusion/pkg/log"
	models2 "kusionstack.io/kusion/pkg/models"
)

const MaxLogLength = 3751

func KCLResult2Spec(kclResults []kcl.KCLResult) (*models2.Spec, error) {
	resources := make([]models2.Resource, len(kclResults))

	for i, result := range kclResults {
		// Marshal kcl result to bytes
		bytes, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		msg := string(bytes)
		if len(msg) > MaxLogLength {
			msg = msg[0:MaxLogLength]
		}

		log.Infof("convert kcl result to resource: %s", msg)

		// Parse json data as models.Resource
		var item models2.Resource
		if err = json.Unmarshal(bytes, &item); err != nil {
			return nil, err
		}
		resources[i] = item
	}

	return &models2.Spec{Resources: resources}, nil
}
