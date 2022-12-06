// Package terraform contains codes about "terraform/dag" and "terraform/tfdiags" copied from Terraform.
//
// Terraform is an open-source project but makes almost all codes in an internal package. It's not a friendly way to benefit the open-source community.
// The latest non-internal Terraform version is v0.15.3, but the k8s related modules in it are less than v1.15. That is too old for us
// and is incompatible with our existing modules.
//
// We import terraform/dag to reuse its DAG structure, and we will build our own DAG structure to support more complex Kusion workflow later.
// We import terraform/tfdiags because terraform/dag imported this module, this module will be deleted together when we replace the DAG
// structure with our own.
package terraform
