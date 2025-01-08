export const RUNS_TYPES = {
  Preview: 'Preview',
  Generate: 'Generate',
  Apply: 'Apply',
  Destroy: 'Destroy',
}

export const RUNS_STATUS_MAP = {
  Scheduling: "Scheduling",
  InProgress: "InProgress",
  Failed: "Failed",
  Succeeded: "Succeeded",
  Cancelled: "Cancelled",
  Queued: "Queued",
}

export const RUNS_ACTION_MAP = {
  Undefined: 'Undefined',
  UnChanged: 'UnChanged',
  Create: 'Create',
  Update: 'Update',
  Delete: 'Delete',
}


export const DEFAULT_WORKSPACE_YAML = `
# modules:
#   \${module_identifier} or \${KAM_name}:
#     path: oci://ghcr.io/kusionstack/module-name # url of the module artifact
#     version: 0.2.0 # version of the module
#     configs: 
#       default: # default configuration, applied to all projects
#         \${field1}: \${value1}
#         #\${field2}: \${value2}
#         ...
#       \${patcher_name}: #patcher configuration, applied to the projects assigned in projectSelector
#         \${field1}: \${value1_override}
#         ...
#         projectSelector:
#         - \${project1_name}
#         - \${project2_name}
#     ...
#
# secretStore: # access the sensitive data stored in a cloud-based secrets manager
#   provider:
#     aws:
#       region: \${region}
#       profile: \${profile}
#     ...
#
# context:
#   \${context_name}: \${context_value}
#   ...
`;
