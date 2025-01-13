import yaml from 'js-yaml'
import _ from 'lodash'
import isUrl from 'is-url'
import { message } from 'antd'

export const yaml2json = (yamlStr: string) => {
  try {
    return {
      data: yaml.load(yamlStr),
      error: false,
    }
  } catch (err) {
    return {
      data: '',
      error: true,
    }
  }
}

export const josn2yaml = (jsonStr: string) => {
  try {
    return {
      data: yaml.dump(JSON.parse(jsonStr)),
      error: false,
    }
  } catch (err) {
    return {
      data: '',
      error: true,
    }
  }
}

export function generateTopologyData(data) {
  const nodes = []
  const edges = []
  const edgeSet = new Set()

  for (const key in data) {
    nodes.push({
      id: key,
      label: key,
      data: {
        count: data[key].count,
        resourceGroup: data[key].resourceGroup,
      },
    })
  }

  function addEdge(source, target) {
    const edgeId = `${source}->${target}`
    if (!edgeSet.has(edgeId)) {
      edges.push({
        source,
        target,
      })
      edgeSet.add(edgeId)
    }
  }

  for (const key in data) {
    const relationships = data[key].relationship
    for (const targetKey in relationships) {
      const relationType = relationships[targetKey]
      if (relationType === 'child') {
        addEdge(key, targetKey)
      } else if (relationType === 'parent') {
        addEdge(targetKey, key)
      }
    }
  }

  return { nodes, edges }
}

export function generateResourceTopologyData(data) {
  const nodes = []
  const edges = []

  const addNode = (id, label, resourceGroup) => {
    nodes.push({ id, label, resourceGroup })
  }

  const uniqueEdges = new Set()
  const addEdge = (source, target) => {
    const edgeKey = `${source}=>${target}`
    if (!uniqueEdges.has(edgeKey)) {
      edges.push({ source, target })
      uniqueEdges.add(edgeKey)
    }
  }

  Object.keys(data).forEach(key => {
    const entity = data[key]

    addNode(key, key.split(':')[1].split('.')[1], entity?.resourceGroup)

    entity.children.forEach(child => {
      addEdge(key, child)
    })

    if (entity.Parents) {
      entity.Parents.forEach(parent => {
        addEdge(parent, key)
      })
    }
  })
  return { nodes, edges }
}

export function getDataType(data) {
  const map = new Map()
  map.set('[object String]', 'String')
  map.set('[object Number]', 'Number')
  map.set('[object Boolean]', 'Boolean')
  map.set('[object Symbol]', 'Symbol')
  map.set('[object Undefined]', 'Undefined')
  map.set('[object Null]', 'Null')
  map.set('[object Function]', 'Function')
  map.set('[object Date]', 'Date')
  map.set('[object Array]', 'Array')
  map.set('[object RegExp]', 'RegExp')
  map.set('[object Error]', 'Error')
  map.set('[object HTMLDocument]', 'HTMLDocument')
  const type = Object.prototype.toString.call(data)
  return map.get(type)
}

export function capitalized(word) {
  return word.charAt(0).toUpperCase() + word?.slice(1)
}

export function isEmptyObject(obj) {
  return obj && Object.keys(obj)?.length === 0
}

export function hasDuplicatesOfObjectArray(array) {
  const seen = new Set()
  for (const item of array) {
    const serialized = JSON.stringify(item)
    if (seen.has(serialized)) {
      return true
    }
    seen.add(serialized)
  }
  return false
}

export function filterKeywordsOfArray(list, keywords, attribute) {
  const result = []
  if (keywords?.length === 1) {
    list?.forEach((item: any) => {
      if (_.get(item, attribute)?.toLowerCase()?.includes(keywords?.[0])) {
        result.push(item)
      }
    })
  } else {
    list?.forEach((item: any) => {
      if (
        keywords?.every((innerValue: string) =>
          _.get(item, attribute)?.toLowerCase()?.includes(innerValue),
        )
      ) {
        result.push(item)
      }
    })
  }
  return result
}

export function getTextSizeByCanvas(
  text: string,
  fontSize: number,
  fontFamily?: string,
) {
  const canvas = document.createElement('canvas')
  const context = canvas.getContext('2d')
  context.font = `${fontSize}px ${fontFamily}`
  const metrics = context.measureText(text)
  const width = Math.ceil(metrics.width)
  canvas.remove()
  return width + 2
}

export function getHistoryList(key) {
  return localStorage?.getItem(key)
    ? JSON.parse(localStorage?.getItem(key))
    : []
}

export function deleteHistoryByItem(key, val: string) {
  const lastHistory: any = localStorage.getItem(key)
  const tmp = lastHistory ? JSON.parse(lastHistory) : []
  if (tmp?.length > 0 && tmp?.includes(val)) {
    const newList = tmp?.filter(item => item !== val)
    localStorage.setItem(key, JSON.stringify(newList))
  }
}

export function cacheHistory(key, val: string) {
  const lastHistory: any = localStorage.getItem(key)
  const tmp = lastHistory ? JSON.parse(lastHistory) : []
  const newList = [val, ...tmp?.filter(item => item !== val)]
  localStorage.setItem(key, JSON.stringify(newList))
  return getHistoryList(key)
}

export function validatorIsUrl(urlString) {
  return isUrl(urlString)
}

export const mockYaml = `apiVersion: cluster.karpor.io/v1beta1\nkind: Cluster\nmetadata:\n  creationTimestamp: \"2024-06-18T08:29:50Z\"\n  managedFields:\n  - apiVersion: cluster.karpor.io/v1beta1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:spec:\n        f:access:\n          f:caBundle: {}\n          f:credential:\n            .: {}\n            f:type: {}\n            f:x509:\n              .: {}\n              f:certificate: {}\n              f:privateKey: {}\n          f:endpoint: {}\n        f:description: {}\n        f:displayName: {}\n    manager: karpor\n    operation: Update\n    time: \"2024-06-18T08:29:50Z\"\n  name: ack\n  resourceVersion: \"40719\"\n  uid: 551a73ce-d604-457c-8ff3-436e0f20c88b\nspec:\n  access:\n    caBundle: NDRjNyoqKioqKioqKioqKioqKioqKioqKioqKjQ0YTg=\n    credential:\n      type: X509Certificate\n      x509:\n        certificate: NTYwZioqKioqKioqKioqKioqKioqKioqKioqKmU3N2M=\n        privateKey: NDVhZSoqKioqKioqKioqKioqKioqKioqKioqKjI1ZmE=\n    endpoint: https://172.24.73.170:32760\n  description: This is a mock ack cluster for demo.\n  displayName: ack\n  provider: \"\"\nstatus: {}\n`

export const mockNewYaml = `apiVersion: cluster.karpor.io/v1beta1111\nkind: Cluster\nmetadata:\n  creationTimestamp: \"2024-06-18T08:59:50Z\"\n  managedFields:\n  - apiVersion: cluster.karpor.io/v1beta1\n    fieldsType: FieldsV12\n    fieldsV1:\n      f:spec:\n        f:access:\n          f:caBundle: {}\n          f:credential:\n            .: {}\n            f:type: {}\n            f:x509:\n              .: {}\n              f:certificate: {}\n              f:privateKey: {}\n          f:endpoint: {}\n        f:description: {}\n        f:displayName: {}\n    manager: karpor2222\n    operation: Update\n    time: \"2024-06-18T08:29:50Z\"\n  name: ack\n  resourceVersion: \"40719\"\n  uid: 551a73ce-d604-457c-8ff3-436e0f20c88b\nspec:\n  access:\n    caBundle: NDRjNyoqKioqKioqKioqKioqKioqKioqKioqKjQ0YTg=\n    credential:\n      type: X509Certificate\n      x509:\n        certificate: NTYwZioqKioqKioqKioqKioqKioqKioqKioqKmU3N2M=\n        privateKey: NDVhZSoqKioqKioqKioqKioqKioqKioqKioqKjI1ZmE=\n    endpoint: https://172.24.73.170:32760\n  description: This is a mock ack cluster for demo.\n  displayName: ack\n  provider: \"\"\nstatus: {}\n`


export const topologyData = {
  "resources": {
    "apps/v1:Deployment:test-project-1112:test-project-1112-dev-helloworld": {
      "resourceType": "apps/v1/Deployment",
      "resourcePlane": "Kubernetes",
      "resourceName": "test-project-1112/test-project-1112-dev-helloworld",
      "iamResourceID": "",
      "cloudResourceID": "",
      "status": "applied"
    },
    "v1:Namespace:test-project-1112": {
      "resourceType": "v1/Namespace",
      "resourcePlane": "Kubernetes",
      "resourceName": "test-project-1112",
      "iamResourceID": "",
      "cloudResourceID": "",
      "status": "applied"
    },
    "v1:Service:test-project-1112:test-project-1112-dev-helloworld-private": {
      "resourceType": "v1/Service",
      "resourcePlane": "Kubernetes",
      "resourceName": "test-project-1112/test-project-1112-dev-helloworld-private",
      "iamResourceID": "",
      "cloudResourceID": "",
      "status": "applied"
    }
  },
  "relations": [
    {
      "DependentResource": "v1:Namespace:test-project-1112",
      "DependencyResource": "apps/v1:Deployment:test-project-1112:test-project-1112-dev-helloworld"
    },
    {
      "DependentResource": "v1:Service:test-project-1112:test-project-1112-dev-helloworld-private",
      "DependencyResource": "apps/v1:Deployment:test-project-1112:test-project-1112-dev-helloworld"
    }
  ],
  "workload": "apps/v1:Deployment:test-project-1112:test-project-1112-dev-helloworld"
}

export function generateG6GraphData(data) {
  const nodes = [];
  const edges = [];
  Object?.entries(data?.resources)?.forEach(([key, value]: any) => {
    nodes.push({
      id: key,
      label: value?.resourceName,
      nodeData: value,
    })
  })
  data?.relations?.forEach(item => {
    edges.push({
      target: item?.DependencyResource,
      source: item?.DependentResource
    })
  })
  return {
    nodes,
    edges
  }
}

export function copy(valStr) {
  const textarea = document.createElement('textarea')
  textarea.value = valStr
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  message.success("Copy Successful")
  document.body.removeChild(textarea)
}