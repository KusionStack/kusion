import React, { useEffect, useRef, useState } from 'react'
import styles from "./styles.module.less"
import { Button, Card, Col, DatePicker, Form, Input, Row, Space, Table, Tabs, Tag } from 'antd'
import { ArrowLeftOutlined, CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { ResourceService } from '@kusionstack/kusion-api-client-sdk'
import G6Tree from '@/components/g6Tree'

const ResourceGraph = () => {

  async function getResourceGraph() {
    // const response = await ResourceService.getResourceGraph({ query: 'sele' });
    // console.log(response, "=====ResourceService.getResourceGraph=====")
  }

  useEffect(() => {
    getResourceGraph()
  }, [])

  return (
    <div className={styles.project_graph}>
      <G6Tree />
    </div>
  )
}

export default ResourceGraph
