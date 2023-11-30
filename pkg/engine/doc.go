// Package engine holds code that drive the engine of Kusion.
//
// Kusion Engine is a middle layer between Intent and the actual infrastructure.
// The major function of this engine is to parse the Intent and to turn all actual infra resources into the desired state
// described in the Intent and almost all operations from the command line will invoke this engine to finish their jobs.
//
// It consists of 3 parts:
//
// 1. Operation Engine: this part is the entrypoint of the whole Kusion Engine and is responsible for kusion basic operations like Preview,
// Apply, Destroy, etc. The main workflow of this part is to parse resources in the Intent, figure out which resource should be modified
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
//	                 +-------------+
//	                 |  Operation  |
//	                 |  Preview    |
//	                 +-------------+
//	                        |
//	                        |
//	                 +-------------+
//	                 |  Operation  |
//	       +---------|  Engine     |----------+
//	       |         +-------------+          |
//	       |                                  |
//	+-------------+                    +-------------+
//	|   State     |                    |   Runtime   |
//	+-------------+                    +-------------+
//
// 1. `Operation Engine` parses resources in Intent and converts them into a DAG
//
// 2. Walk this DAG:
//  1. Get the latest `State` from the actual infra by the `Runtime`
//  2. Get the last operation `State` from the `State` storage medium
//
// 3. Merge/Diff three states: desired state described in Intent, live state from `Runtime` and prior state from `State` storage medium, and return the diff result to the console
package engine
