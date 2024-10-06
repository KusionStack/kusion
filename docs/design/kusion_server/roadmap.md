# Kusion Server Roadmap

## Current gaps between upstream and internal version

The following can be contributed back to upstream in August with the exception of `KCL Configuration Management`.

### KCL Configuration Management
- Initialize and modify `.k` files with new branches
- Update configuration items
- Manage configuration versions

### AuthN, AuthZ
- Support authentication based on JWT token
  - Support auth whitelist
  - Support RBAC if an IAM system is present
  - Support ability to generate a token if an IAM system is present

### Version management
- Stack status management
  - Persistent last `generated`/`previewed`/`applied` spec version
  - Support explicit spec ID or content when `previewing`/`applying`

### Resources management
- Centrally write resource metadata and attributes after apply
- Basic ability to look them up based on different criteria:
  - Project/Stack/Workspace
    - Example `"All resources for project X, stack Y in workspace Z"`
  - ResourcePlane/ResourceType
    - Example: `"All AWS S3 buckets for project X"`
  - Any field in resource attributes
    - Example: `"All MySQL databases in production with version 5.7 in region us-west-1"`

### Secrets management
- Manage application secrets with third-party vendors
  - API to manage secrets associated with a `stack` (usually application secrets)
  - Automatically replace them to actual values during preview/apply without showing them
- Workspace credential management with third-party vendors
  - API to manage `KubeConfig`/`AccessKey`/`Secretkey` associated with a `workspace`
  - Automatically replace them to actual values during preview/apply without showing them

### Quality-of-life improvements
- Support applying through a federated cluster via a gateway runtime
- The ability to list `projects`/`stacks`/`workspaces` with custom filters, for example:
  - list all `projects` under org x
  - list all `stacks` under org x and project y
  - find `workspace` with the name x
- Better search experience
  - Project name and path validation
  - The ability to point to a `project` by name instead of ID
  - The ability to automatically create relevant orgs when creating projects
    - if exists, use it; otherwise create a new one
- The ability to set up a default backend during startup
- Automatically lock the `stack` when other operations are in progress
  - Gating mechanism for proceeding into the next stage in lifecycle (generate -> preview -> apply)
- Automatically generate swag docs

## Next up...

These are what's considered interesting in the next few months for Kusion server to work on...

### AuthN and AuthZ
- This will likely be integration with open-source IAM solutions
- More ways to AuthN and AuthZ than JWT
- Support organizational roles and role assignments

### Resource Graphs
- Grant customized permissions on resources created by Kusion server
- The ability to render a topological view of resources on said dimensions

### GitOps
- Manage automated GitOps triggers
- Support master-worker architecture

### Variables and Variable Sets
- Support platform configurations that can set on any scopes: global, org, project, workspace, etc
- Define priorities when multiple variables with the same key are present and may result in a scope conflict
- Support more third-party secret management solutions for storing sensitive variables 

### Audit trails and workflows
- Create audit trails with subject, object and operation
- Persist workflow executions, e.g. all the generate, preview, apply operations
- The ability to insert custom stages in these workflow, e.g. approvals

### Policies
- Render-time (generate) policies and Execution-time (preview/apply) policy
- Both built-in and customizable