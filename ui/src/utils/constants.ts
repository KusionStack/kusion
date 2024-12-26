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
