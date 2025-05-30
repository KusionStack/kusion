package operation

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/copier"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/parser"
	"kusionstack.io/kusion/pkg/engine/release"
	resourcegraph "kusionstack.io/kusion/pkg/engine/resource/graph"
	runtimeinit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/third_party/terraform/dag"
	"kusionstack.io/kusion/third_party/terraform/tfdiags"
)

type ApplyOperation struct {
	models.Operation
}

type ApplyRequest struct {
	models.Request
	Release *apiv1.Release
	Graph   *apiv1.Graph
}

type ApplyResponse struct {
	Release *apiv1.Release
	Graph   *apiv1.Graph
}

// Apply means turn all actual infra resources into the desired state described in the request by invoking a specified Runtime.
// Like other operations, Apply has 3 main steps during the whole process.
//  1. parse resources and their relationship to build a DAG and should take care of those resources that will be deleted
//  2. walk this DAG and execute all graph nodes concurrently, besides the entire process should follow dependencies in this DAG
//  3. during the execution of each node, it will invoke different runtime according to the resource type
func (ao *ApplyOperation) Apply(req *ApplyRequest) (rsp *ApplyResponse, s v1.Status) {
	log.Infof("engine: Apply start!")
	o := ao.Operation

	// keep release state
	rsp = &ApplyResponse{
		Release: req.Release,
	}

	defer func() {
		close(o.MsgCh)

		if e := recover(); e != nil {
			log.Error("apply panic:%v", e)

			switch x := e.(type) {
			case string:
				s = v1.NewErrorStatus(fmt.Errorf("apply panic:%s", e))
			case error:
				s = v1.NewErrorStatus(x)
			default:
				s = v1.NewErrorStatusWithCode(v1.Unknown, errors.New("unknown panic"))
			}
		}
	}()

	if s = validateApplyRequest(req); v1.IsErr(s) {
		return rsp, s
	}

	// Update the operation semaphore.
	if err := o.UpdateSemaphore(); err != nil {
		return rsp, v1.NewErrorStatus(err)
	}

	// 1. init & build Indexes
	priorState := req.Release.State
	priorStateResourceIndex := priorState.Resources.Index()
	// copy priorStateResourceIndex into a new map
	stateResourceIndex := map[string]*apiv1.Resource{}
	for k, v := range priorStateResourceIndex {
		stateResourceIndex[k] = v
	}

	runtimesMap, s := runtimeinit.Runtimes(*req.Release.Spec, *req.Release.State)
	if v1.IsErr(s) {
		return rsp, s
	}
	o.RuntimeMap = runtimesMap

	// 2. build & walk DAG
	applyGraph, s := newApplyGraph(req.Release.Spec, priorState)
	if v1.IsErr(s) {
		return rsp, s
	}
	log.Infof("Apply Graph:\n%s", applyGraph.String())
	// Get dependencies and dependents of each node to be populated into resource graph.
	resourceGraph := populateResourceGraph(applyGraph, req.Graph)

	rel, s := copyRelease(req.Release)
	if v1.IsErr(s) {
		return rsp, s
	}
	applyOperation := &ApplyOperation{
		Operation: models.Operation{
			OperationType:           models.Apply,
			ReleaseStorage:          o.ReleaseStorage,
			SecretStore:             req.Release.Spec.SecretStore,
			CtxResourceIndex:        map[string]*apiv1.Resource{},
			PriorStateResourceIndex: priorStateResourceIndex,
			StateResourceIndex:      stateResourceIndex,
			RuntimeMap:              o.RuntimeMap,
			Stack:                   o.Stack,
			IgnoreFields:            o.IgnoreFields,
			MsgCh:                   o.MsgCh,
			WatchCh:                 o.WatchCh,
			Lock:                    &sync.Mutex{},
			Release:                 rel,
			Sem:                     o.Sem,
		},
	}

	w := &dag.Walker{Callback: applyOperation.walkFun}
	w.Update(applyGraph)
	// Wait
	if diags := w.Wait(); diags.HasErrors() {
		s = v1.NewErrorStatus(diags.Err())
	}
	rsp.Release = applyOperation.Release
	rsp.Graph = resourceGraph
	return rsp, s
}

func (ao *ApplyOperation) walkFun(v dag.Vertex) (diags tfdiags.Diagnostics) {
	return applyWalkFun(&ao.Operation, v)
}

func validateApplyRequest(req *ApplyRequest) v1.Status {
	if req == nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "request is nil")
	}
	if err := release.ValidateRelease(req.Release); err != nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, err.Error())
	}
	if err := resourcegraph.ValidateGraph(req.Graph); err != nil {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, err.Error())
	}
	if req.Release.Phase != apiv1.ReleasePhaseApplying {
		return v1.NewErrorStatusWithMsg(v1.InvalidArgument, "release phase is not applying")
	}
	return nil
}

func newApplyGraph(spec *apiv1.Spec, priorState *apiv1.State) (*dag.AcyclicGraph, v1.Status) {
	specParser := parser.NewIntentParser(spec)
	g := &dag.AcyclicGraph{}
	g.Add(&graph.RootNode{})

	s := specParser.Parse(g)
	if v1.IsErr(s) {
		return nil, s
	}
	deleteResourceParser := parser.NewDeleteResourceParser(priorState.Resources)
	s = deleteResourceParser.Parse(g)
	if v1.IsErr(s) {
		return nil, s
	}

	return g, s
}

func copyRelease(r *apiv1.Release) (*apiv1.Release, v1.Status) {
	rel := &apiv1.Release{}
	if err := copier.Copy(rel, r); err != nil {
		return nil, v1.NewErrorStatusWithMsg(v1.Internal, fmt.Sprintf("copy release failed, %v", err))
	}
	return rel, nil
}

func applyWalkFun(o *models.Operation, v dag.Vertex) (diags tfdiags.Diagnostics) {
	var s v1.Status
	if v == nil {
		return nil
	}

	if o.Sem != nil {
		o.Sem.Acquire()
		defer o.Sem.Release()
	}

	if node, ok := v.(graph.ExecutableNode); ok {
		if rn, ok2 := v.(*graph.ResourceNode); ok2 {
			o.MsgCh <- models.Message{ResourceID: rn.Hashcode().(string)}

			s = node.Execute(o)
			if v1.IsErr(s) {
				o.MsgCh <- models.Message{
					ResourceID: rn.Hashcode().(string), OpResult: models.Failed,
					OpErr: fmt.Errorf("node execte failed, status:\n%v", s),
				}
			} else {
				o.MsgCh <- models.Message{ResourceID: rn.Hashcode().(string), OpResult: models.Success}
			}
		} else {
			s = node.Execute(o)
		}
	}
	if s != nil {
		diags = diags.Append(fmt.Errorf("apply failed, status:\n%v", s))
	}
	return diags
}

// populateResourceGraph populate dependents and dependencies of each resource in resource graph with acyclicGraph
func populateResourceGraph(applyGraph *dag.AcyclicGraph, resourceGraph *apiv1.Graph) *apiv1.Graph {
	for _, vertex := range applyGraph.Vertices() {
		// Get resource ID from vertex.
		resourceName := dag.VertexName(vertex)
		resource := resourcegraph.FindGraphResourceByID(resourceGraph.Resources, resourceName)
		// Populate it's dependents and dependencies
		if resource != nil {
			dependents := applyGraph.DownEdges(vertex)
			dependencies := applyGraph.UpEdges(vertex)

			for _, dependent := range dependents {
				dependentName := dag.VertexName(dependent)
				resource.Dependents = append(resource.Dependents, dependentName)
			}

			for _, dependency := range dependencies {
				dependencyName := dag.VertexName(dependency)
				resource.Dependencies = append(resource.Dependencies, dependencyName)
			}
		}
	}

	return resourceGraph
}
