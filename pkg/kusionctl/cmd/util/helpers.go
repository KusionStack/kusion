package util

import (
	"errors"
	"fmt"

	"kusionstack.io/kusion/pkg/engine/models"
	runtimeInit "kusionstack.io/kusion/pkg/engine/runtime/init"
)

func RecoverErr(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = errors.New(x)
		case error:
			*err = x
		default:
			*err = errors.New("unknow panic")
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ValidateResourceType(runtimes map[models.Type]runtimeInit.InitFn, t models.Type) (runtimeInit.InitFn, error) {
	supportTypes := make([]models.Type, 0, len(runtimes))
	for k := range runtimes {
		supportTypes = append(supportTypes, k)
	}
	typeFun := runtimes[t]
	if typeFun == nil {
		return nil, fmt.Errorf("unknow resource type: %s. Currently supported resource types are: %v", t, supportTypes)
	}
	return typeFun, nil
}
