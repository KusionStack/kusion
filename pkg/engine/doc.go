// Package engine holds code that drive the engine of Kusion.
//
// Kusion Engine is a middle layer between Spec and the actual infrastructure.
// The major function of this engine is to parse the Spec and to turn all actual infra resources into the desired state
// described in the Spec and almost all operations from the command line will invoke this engine to finish their jobs.
//
// It consists of 3 parts:
//
// 1. Operation Engine: this part is the entrypoint of the whole Kusion Engine and is responsible for kusion basic operations like Preview,
// Apply, Destroy, etc. The main workflow of this part is to parse resources in the Spec, figure out which resource should be modified
// according to specified operation type, and execute this operation to the real infra resources. During this workflow,
// the following two parts will be involved.
//
// 2. Runtime: it is an interface between the actual infrastructure and Kusion. All operations that trying to manipulate a resource
// should be delegated to one Runtime to make this operation effect in the actual infrastructure
//
// 3. State: state is a record of an operation's result. It is often used as a datasource for 3-way merge/diff in operations like Apply or Preview
//
// Let's get operation Preview as an example to demonstrate how the three parts cooperate in an actual operation.
//
//                        +-------------+
//                        |  Operation  |
//                        |  Preview    |
//                        +-------------+
//                               |
//                               |
//                        +-------------+
//                        |  Operation  |
//              +---------|  Engine     |----------+
//              |         +-------------+          |
//              |                                  |
//       +-------------+                    +-------------+
//       |   State     |                    |   Runtime   |
//       +-------------+                    +-------------+
//
// 1. parse resources in the Spec and convert into a DAG
//
// 2. Walk this DAG:
// 	   a) get the latest state from the actual infra by the Runtime
//     b) get last operation state from the State
//
// 3. Diff the two states(live state and prior state) and return the details of these diffs to cmd
package engine
