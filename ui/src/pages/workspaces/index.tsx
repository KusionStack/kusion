import React, { useEffect, useState } from 'react'
import { Button, Col, Form, Input, message, Row, Select, Space } from 'antd'
import {
  PlusOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom';
import { BackendService, WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import WorkspaceCard from './components/workspaceCard';
import WorkscpaceForm from './components/workscpaceForm';

import styles from './styles.module.less'

const Workspaces = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm();
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [open, setOpen] = useState(false)
  const [searchParams, setSearchParams] = useState({
    pageSize: 30,
    page: 1,
    query: undefined,
    total: undefined,
  })
  const [workspaceList, setWorkspaceList] = useState([]);
  const [backendList, setBackendList] = useState([]);


  async function getBackendList() {
    try {
      const response: any = await BackendService.listBackend({
        query: {
          page: 1,
          pageSize: 10000,
        }
      });
      if (response?.data?.success) {
        setBackendList(response?.data?.data?.backends);
      }
    } catch (error) {

    }
  }


  async function getListWorkspace(params) {
    try {
      const response: any = await WorkspaceService.listWorkspace({
        query: {
          backendID: params?.query?.backendID,
          page: params?.page || searchParams?.page,
          pageSize: params?.pageSize || searchParams?.pageSize,
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
    getBackendList()
    getListWorkspace(searchParams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleAdd() {
    setActionType('ADD')
    setOpen(true)
  }

  function handleEdit(record) {
    setActionType('EDIT')
    setOpen(true)
    setFormData(record)
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
    let response: any
    if (actionType === 'EDIT') {
      response = await WorkspaceService.updateWorkspace({
        body: {
          ...values,
          // TODO: get actual owner
          owners: ['default']
        },
        path: {
          workspaceID: (formData as any)?.id
        }
      })
    } else {
      response = await WorkspaceService.createWorkspace({
        body: {
          ...values,
          // TODO: get actual owner
          owners: ['default']
        }
      })
    }
    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getListWorkspace(searchParams)
      setOpen(false)
      if (actionType === 'ADD') {
        navigate(`/workspaces/detail/${response?.data?.data?.id}`)
      }
    } else {
      message.error(response?.data?.message)
    }
  }

  function handleClose() {
    setFormData(undefined)
    setOpen(false)
  }

  const arrayColByN = conversionArray(workspaceList, 4)

  async function confirmDelete(record) {
    const response: any = await WorkspaceService.deleteWorkspace({
      path: {
        workspaceID: record?.id,
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getListWorkspace(searchParams)
    }
  }

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
            <Form.Item name="backendID" label="Backend">
              <Select style={{ width: 200 }}>
                {
                  backendList?.map(item => <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>)
                }
              </Select>
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
                    <WorkspaceCard
                      title={innerItem?.name}
                      desc={innerItem?.description}
                      createDate={innerItem?.creationTimestamp}
                      nickName={innerItem?.owners}
                      onClick={() => navigate(`/workspaces/detail/${innerItem?.id}?workspaceName=${innerItem?.name}`)}
                      onDelete={() => confirmDelete(innerItem)}
                      handleEdit={() => handleEdit(innerItem)}
                    />
                  </Col>
                )
              })}
            </Row>
          )
        })}
      </div>
      <WorkscpaceForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} actionType={actionType} formData={formData} />
    </div>
  )
}

export default Workspaces
