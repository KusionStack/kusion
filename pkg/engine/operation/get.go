package operation

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/status"
)

type GetOperation struct {
	opsmodels.Operation
}

type GetRequest struct {
	opsmodels.Request `json:",inline" yaml:",inline"`
}

type GetResponse struct {
	// TODO:
	Order *opsmodels.ChangeOrder
}

func (o *GetOperation) Get(request *GetRequest) error {
	fmt.Println("inside get")

	// defer func() {
	// 	if e := recover(); e != nil {
	// 		log.Error("get panic:%v", e)

	// 		switch x := e.(type) {
	// 		case string:
	// 			s = status.NewErrorStatus(fmt.Errorf("get panic:%s", e))
	// 		case error:
	// 			s = status.NewErrorStatus(x)
	// 		default:
	// 			s = status.NewErrorStatus(errors.New("unknown panic"))
	// 		}
	// 	}
	// }()

	if s := validateRequest(&request.Request); status.IsErr(s) {
		return fmt.Errorf(s.Message())
	}

	fmt.Println("validate request")

	// 1. init & build Indexes
	priorState, _ := o.InitStates(&request.Request)

	fmt.Println("after init & build indexes")

	// Kusion is a multi-runtime system. We initialize runtimes dynamically by resource types
	resources := request.Spec.Resources
	// resources = append(resources, priorState.Resources...)
	runtimesMap, s := runtimeinit.Runtimes(resources)
	if status.IsErr(s) {
		return fmt.Errorf(s.Message())
	}
	o.RuntimeMap = runtimesMap
	o.PriorStateResourceIndex = priorState.Resources.Index()

	fmt.Println("after initialize runtimes")

	readResponseRet := make(map[string]*runtime.ReadResponse, resources.Len())
	// Keep sorted
	ids := make([]string, resources.Len())
	for i := range resources {
		res := &resources[i]
		t := res.Type

		// Save id first, might have TF resources
		ids[i] = res.ResourceKey()

		// get prior resource which is stored in kusion_state.json
		priorResource := o.PriorStateResourceIndex[ids[i]]

		// get the live resource from runtime
		readRequest := &runtime.ReadRequest{
			PlanResource:  res,
			PriorResource: priorResource,
			Stack:         o.Stack,
		}
		resourceType := res.Type

		// print readRequest
		// fmt.Println(jsonutil.Marshal2String(readRequest))
		// fmt.Println(jsonutil.Marshal2String(resourceType))

		response := o.RuntimeMap[resourceType].Read(context.Background(), readRequest)
		// liveResource := response.Resource
		// s := response.Status

		// only for debug

		if response == nil {
			// log.Debug("unsupported resource type: %s", t)
			fmt.Printf("unsupported resource type: %s", t)
			continue
		}

		// add data
		readResponseRet[ids[i]] = response

		// fmt.Println(jsonutil.Marshal2String(response.Resource))

		if status.IsErr(response.Status) {
			return fmt.Errorf(response.Status.String())
		}
	}

	 // Format the table header
    fmt.Printf("%-10s%-15s%-20s%-30s%-10s\n", "ID", "Kind", "Name", "Detail", "Status")
	// tables := make(map[string]*printers.Table, len(ids))
	for i := range ids {
		// how to skip Terraform
		resource := readResponseRet[ids[i]].Resource
		attrs := printers.Convert(&unstructured.Unstructured{Object: resource.Attributes})
		detail, ready := printers.Generate(attrs)
		printTable(ids[i], attrs.GetObjectKind().GroupVersionKind().Kind, resource.ID, detail, ready)
	}
	return nil
}

func printTable(id string, kind string, name string, detail string, status bool) {
    // Print a separator line
    fmt.Println("----------------------------------------------------------")

    // Format and print the data row
	fmt.Println("ID:", id)
    fmt.Println("Kind:", kind)
    fmt.Println("Name:", name)
    fmt.Println("Detail:", detail)
    fmt.Println("Status:", status)
}
