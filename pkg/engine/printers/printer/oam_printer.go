package printer

import (
	"fmt"
	"strconv"
	"strings"

	oamv1beta1 "kusionstack.io/kusion/third_party/kubevela/kubevela/apis/v1beta1"
)

func AddOAMHandlers(h PrintHandler) {
	_ = h.TableHandler(printApplication)
}

func printApplication(obj *oamv1beta1.Application) (string, bool) {
	// Component and Type
	components := obj.Spec.Components
	componentNames := make([]string, 0, len(components))
	componentTypes := make([]string, 0, len(components))
	for i := range components {
		componentNames = append(componentNames, components[i].Name)
		componentTypes = append(componentTypes, components[i].Type)
	}
	var componentStr string
	if len(componentNames) != 0 {
		componentStr = strings.Join(componentNames, ",")
	}
	var typeStr string
	if len(componentTypes) > 0 {
		typeStr = strings.Join(componentTypes, ",")
	}

	// Phase
	phase := obj.Status.Phase

	// Healthy and Status
	services := obj.Status.Services
	serviceHealths := make([]string, 0, len(services))
	serviceStatuses := make([]string, 0, len(services))
	for i := range services {
		serviceHealths = append(serviceHealths, strconv.FormatBool(services[i].Healthy))
		msg := services[i].Message
		if msg == "" {
			msg = "None"
		}
		serviceStatuses = append(serviceStatuses, msg)
	}
	var healthyStr string
	if len(serviceHealths) > 0 {
		healthyStr = strings.Join(serviceHealths, ",")
	}
	var statusStr string
	if len(serviceStatuses) > 0 {
		statusStr = strings.Join(serviceStatuses, ",")
	}

	// Age
	age := translateTimestampSince(obj.CreationTimestamp)

	return fmt.Sprintf("Phase: %s, Component: %s, Type: %s, Healthy: %s, Status: %s, Age: %s",
		phase, componentStr, typeStr, healthyStr, statusStr, age), phase == "running" && obj.Status.Workflow.Finished
}
