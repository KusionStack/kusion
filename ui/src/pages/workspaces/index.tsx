import React, { useEffect, useState } from 'react'
import { Button, Col, Form, Input, message, Row, Space } from 'antd'
import {
  PlusOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom';
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import WorkspaceCard from './components/workspaceCard';
import WorkscpaceForm from './components/workscpaceForm';

import styles from './styles.module.less'

const Workspaces = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false)
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: undefined,
  })
  const [workspaceList, setWorkspaceList] = useState([]);


  async function getListWorkspace(params) {
    try {
      const response: any = await WorkspaceService.listWorkspace({
        ...searchParams,
        query: {
          workspaceName: searchParams?.query?.workspaceName,
        }
      });
      if (response?.data?.success) {
        setWorkspaceList(response?.data?.data?.workspaces);
        setSearchParams({
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        })
      } else {
        message.error(response?.data?.messaage)
      }
    } catch (error) {
      console.error('Error:', error);
    }
  }

  useEffect(() => {
    getListWorkspace(searchParams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined
    })
    getListWorkspace({
      page: 1,
      pageSize: 10,
      query: undefined
    })
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    setSearchParams({
      ...searchParams,
      query: values
    })
    getListWorkspace({
      page: 1,
      pageSize: 10,
      query: values,
    })
  }

  async function handleSubmit(values) {
    const response: any = WorkspaceService.createWorkspace({
      body: {
        ...values
      }
    })
    if (response?.data?.success) {
      message.success("Create Success")
      getListWorkspace(searchParams)
      setOpen(false)
    } else {
      message.error(response?.data?.message)
    }
  }

  function handleClose() {
    setOpen(false)
  }

  const arrayColByN = conversionArray(workspaceList, 4)

  return (
    <div className={styles.kusion_workspace_container}>
      <div className={styles.kusion_workspace_action}>
        <h3>Workspaces</h3>
        <div className={styles.kusion_workspace_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Workspace
          </Button>
        </div>
      </div>
      <div className={styles.kusion_workspace_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="workspaceName" label="Workspace Name">
              <Input />
            </Form.Item>
            <Form.Item style={{ marginLeft: 20 }}>
              <Space>
                <Button onClick={handleReset}>Reset</Button>
                <Button type='primary' onClick={handleSearch}>Search</Button>
              </Space>
            </Form.Item>
          </Space>
        </Form>
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
