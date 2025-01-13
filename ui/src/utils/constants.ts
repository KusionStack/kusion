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

const moduleIdentifier = "myModule";
const kamName = "myKAM";
const moduleName = "module-name";
const moduleVersion = "0.2.0";
const field1 = "apiUrl";
const value1 = "https://api.example.com";
const field2 = "timeout";
const value2 = "30s";
const patcherName = "myPatcher";
const field1Override = "https://api.override.com";
const project1Name = "project1";
const project2Name = "project2";
const region = "us-west-2";
const profile = "default";
const contextName = "env";
const contextValue = "production";

export const DEFAULT_WORKSPACE_YAML = `
#  modules:
#    ${moduleIdentifier} or ${kamName}:
#      path: oci://ghcr.io/kusionstack/${moduleName}
#      version: ${moduleVersion}
#      configs: 
#        default:
#          ${field1}: ${value1}
#          ${field2}: ${value2}
#        ${patcherName}:
#          ${field1}: ${field1Override}
#          projectSelector:
#          - ${project1Name}
#          - ${project2Name}
#  secretStore:
#    provider:
#      aws:
#        region: ${region}
#        profile: ${profile}
#  context:
#    ${contextName}: ${contextValue}
`;
