import React, { useEffect, useState } from 'react'
import { message } from 'antd'
import { ResourceService } from '@kusionstack/kusion-api-client-sdk'
import TopologyMap from '@/components/topologyMap'
import { generateG6GraphData } from '@/utils/tools'

import styles from "./styles.module.less"


const ResourceGraph = () => {

  const [graphData, setGraphData] = useState()

  async function getResourceGraph() {
    const response: any = await ResourceService.getResourceGraph({
      query: {
        stack_id: 1
      }
    });
    if (response?.data?.success) {
      setGraphData(response?.data?.data)
    } else {
      message.error(response?.data?.message)
    }
  }

  useEffect(() => {
    getResourceGraph()
  }, [])

  const topologyData = graphData && generateG6GraphData(graphData)

  return (
    <div className={styles.project_graph}>
      <TopologyMap topologyData={topologyData} topologyLoading={false} />
    </div>
  )
}

export default ResourceGraph
