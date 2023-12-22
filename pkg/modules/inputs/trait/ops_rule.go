package trait

import (
	"errors"
	"strconv"

	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

type OpsRule struct {
	MaxUnavailable string `json:"maxUnavailable,omitempty" yaml:"maxUnavailable,omitempty"`
}

const (
	OpsRuleConst        = "opsRule"
	MaxUnavailableConst = "maxUnavailable"
)

func GetMaxUnavailable(opsRule *OpsRule, modulesConfig map[string]apiv1.GenericConfig) (intstr.IntOrString, error) {
	var maxUnavailable intstr.IntOrString
	if opsRule != nil {
		maxUnavailable = intstr.Parse(opsRule.MaxUnavailable)
	} else {
		// An example of opsRule config in modulesConfig
		// opsRule:
		//   maxUnavailable: 1 # or 10%
		if modulesConfig[OpsRuleConst] == nil || modulesConfig[OpsRuleConst][MaxUnavailableConst] == nil {
			return intstr.IntOrString{}, nil
		}
		var wsValue string
		wsValue, isString := modulesConfig[OpsRuleConst][MaxUnavailableConst].(string)
		if !isString {
			temp, isInt := modulesConfig[OpsRuleConst][MaxUnavailableConst].(int)
			if isInt {
				wsValue = strconv.Itoa(temp)
			} else {
				return intstr.IntOrString{}, errors.New("illegal workspace config. opsRule.maxUnavailable in the workspace config is not string or int")
			}
		}
		maxUnavailable = intstr.Parse(wsValue)
	}
	return maxUnavailable, nil
}
