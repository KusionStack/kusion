import { StackService } from "@kusionstack/kusion-api-client-sdk"


export async function createApply(values, stackId) {
  const response: any = await StackService.applyStackAsync({
    body: {
      ...values,
      stackID: Number(stackId),
      workspace: values?.workspace,
    },
    query: {
      workspace: values?.workspace,
      noCache: true,
    },
    path: {
      stackID: Number(stackId),
    }
  })
  return response
}

export async function createGenerate(values, stackId) {
  const response: any = await StackService.generateStackAsync({
    body: {
      ...values,
      stackID: Number(stackId),
      workspace: values?.workspace,
    },
    query: {
      workspace: values?.workspace,
      noCache: true,
    },
    path: {
      stackID: Number(stackId),
    }
  })
  return response
}

export async function createDestroy(values, stackId) {
  const response: any = await StackService.destroyStackAsync({
    body: {
      ...values,
      stackID: Number(stackId),
      workspace: values?.workspace,
    },
    query: {
      workspace: values?.workspace,
      noCache: true,
    },
    path: {
      stackID: Number(stackId),
    }
  })
  return response
}

export async function createPreview(values, stackId) {
  const response: any = await StackService.previewStackAsync({
    body: {
      ...values,
      stackID: Number(stackId),
      workspace: values?.workspace,
    },
    query: {
      workspace: values?.workspace,
      output: 'json',
      noCache: true,
    },
    path: {
      stackID: Number(stackId),
    }
  })
  return response
}

export async function queryListRun(query) {
  return await StackService.listRun({
    query
  });
}
