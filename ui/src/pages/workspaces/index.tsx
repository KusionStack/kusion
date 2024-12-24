import React, { useEffect, useState } from 'react'
import { Button, Card, Col, Input, message, Popconfirm, Row, Tooltip } from 'antd'
import {
  PlusOutlined,
  SortDescendingOutlined,
  SortAscendingOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { BackendService, WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import WorkspaceCard from './components/workspaceCard';
import WorkscpaceForm from './components/workscpaceForm';

import styles from './styles.module.less'
import { useNavigate } from 'react-router-dom';

const orderIconStyle: React.CSSProperties = {
  marginLeft: 0,
}

const Workspaces = () => {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState<string>('')

  const [sortParams, setSortParams] = useState<any>({
    orderBy: 'name',
    isAsc: true,
  })
  const [workspaceList, setWorkspaceList] = useState([]);


  async function getListWorkspace() {
    try {
      const response: any = await WorkspaceService.listWorkspace();
      if (response?.data?.success) {
        setWorkspaceList(response?.data?.data);
      } else {
        message.error("请求失败")
      }
    } catch (error) {
      console.error('Error:', error);
    }
  }

  useEffect(() => {
    getListWorkspace()
  }, [])

  function handleAdd() {
    setOpen(true)
  }


  function conversionArray(baseArray, n) {
    const len = baseArray.length
    const lineNum = len % n === 0 ? len / n : Math.floor(len / n + 1)
    const res = []
    for (let i = 0; i < lineNum; i++) {
      const temp = baseArray.slice(i * n, i * n + n)
      res.push(temp)
    }
    return res
  }

  function handleChange(event) {
    setKeyword(event?.target.value)
  }

  async function handleSubmit(values) {
    console.log(values, "=====handleSubmit values=====")
    const response: any = WorkspaceService.createWorkspace({
      body: {
        ...values
      }
    })
    if (response?.data?.success) {
      message.success("Create Success")
      setOpen(false)
    } else {
      message.error(response?.data?.message || 'Request Faild')
    }
  }

  function handleClose() {
    setOpen(false)
  }

  const arrayColByN = conversionArray(workspaceList, 4)

  return (
    <div className={styles.kusion_workspace_container}>
      <div className={styles.kusion_workspace_toolbar}>
        <Button type="primary" onClick={handleAdd}>
          <PlusOutlined /> New Workspace
        </Button>
        <div className={styles.kusion_workspace_toolbar_right}>
          <Input
            placeholder={'关键字搜索'}
            suffix={<SearchOutlined />}
            style={{ width: 260 }}
            value={keyword}
            onChange={handleChange}
            allowClear
          />
        </div>
      </div>
      <div className={styles.kusion_workspace_content}>
        {arrayColByN?.map((item, index) => {
          return (
            <Row
              key={index}
              gutter={{ xs: 8, sm: 16, md: 24, lg: 32 }}
              style={{ marginBottom: 20 }}
            >
              {item?.map((innerItem, innerIndex) => {
                return (
                  <Col key={innerIndex} className="gutter-row" span={6}>
                    <WorkspaceCard title={innerItem?.name} desc={innerItem?.description} createDate={innerItem?.creationTimestamp} nickName={innerItem?.owners} onClick={() => navigate(`/workspaces/detail/${innerItem?.id}`)} />
                  </Col>
                )
              })}
            </Row>
          )
        })}
      </div>
      <WorkscpaceForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} />
    </div>
  )
}

export default Workspaces
