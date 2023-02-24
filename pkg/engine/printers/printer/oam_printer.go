package printer

import (
	"fmt"
	"strconv"
	"strings"

	oamv1beta1 "kusionstack.io/kusion/third_party/kubevela/kubevela/apis/v1beta1"
)

func AddOAMHandlers(h PrintHandler) {
	h.TableHandler(printApplication)
}

func printApplication(obj *oamv1beta1.Application) (string, bool) {
	// Component and Type
	components := obj.Spec.Components
	componentNames := make([]string, len(components))
	componentTypes := make([]string, len(components))
	for i := range components {
		componentNames[i] = components[i].Name
		componentTypes[i] = components[i].Type
	}
	componentStr := strings.Join(componentNames, ",")
	typeStr := strings.Join(componentTypes, ",")

	// Phase
	phase := obj.Status.Phase

	// Healthy and Status
	services := obj.Status.Services
	serviceHealths := make([]string, len(services))
	serviceStatuses := make([]string, len(services))
	for i := range services {
		serviceHealths[i] = strconv.FormatBool(services[i].Healthy)
		serviceStatuses[i] = services[i].Message
	}
	healthyStr := strings.Join(serviceHealths, ",")
	statusStr := strings.Join(serviceStatuses, ",")

	// Age
	age := translateTimestampSince(obj.CreationTimestamp)

	return fmt.Sprintf("Component: %s, Type: %s, Phase: %s, Healthy: %s, Status: %s, Age: %s",
		componentStr, typeStr, phase, healthyStr, statusStr, age), phase == "running" && obj.Status.Workflow.Finished
}
