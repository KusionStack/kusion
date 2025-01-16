import React, { useEffect, useRef, useState } from 'react'
import { message } from 'antd'
import { ResourceService } from '@kusionstack/kusion-api-client-sdk'
import TopologyMap from '@/components/topologyMap'
import { generateG6GraphData } from '@/utils/tools'

const ResourceGraph = ({ stackId }) => {
  const drawRef = useRef(null)
  const [graphData, setGraphData] = useState()
  const [topologyLoading, setTopologyLoading] = useState(false)

  async function getResourceGraph(id) {
    try {
      setTopologyLoading(true)
      const response: any = await ResourceService.getResourceGraph({
        query: {
          stackID: id
        } as any
      });
      if (response?.data?.success) {
        setGraphData(response?.data?.data)
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {

    } finally {
      setTopologyLoading(false)
    }
  }

  useEffect(() => {
    if (stackId) {
      getResourceGraph(stackId)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stackId])


  useEffect(() => {
    if (graphData) {
      const topologyData = graphData && generateG6GraphData(graphData)
      drawRef.current?.drawGraph(topologyData)
    }
  }, [graphData])

  return (
    <TopologyMap ref={drawRef} topologyLoading={topologyLoading} />
  )
}

export default ResourceGraph
