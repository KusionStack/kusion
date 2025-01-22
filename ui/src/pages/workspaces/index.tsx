import React, { useEffect, useRef, useState } from 'react'
import { Button, Col, Form, message, Pagination, Row, Select, Space } from 'antd'
import {
  PlusOutlined,
} from '@ant-design/icons'
import queryString from 'query-string'
import { useLocation, useNavigate } from 'react-router-dom';
import { BackendService, WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import WorkspaceCard from './components/workspaceCard';
import WorkscpaceForm from './components/workscpaceForm';

import styles from './styles.module.less'

const Workspaces = () => {
  const navigate = useNavigate()
  const location = useLocation();
  const [form] = Form.useForm();
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [open, setOpen] = useState(false)
  const { pageSize = 10, page = 1, total = 0, backendID } = queryString.parse(location?.search);
  const [searchParams, setSearchParams] = useState({
    pageSize,
    page,
    query: {
      backendID
    },
    total,
  });
  const [workspaceList, setWorkspaceList] = useState([]);
  const [backendList, setBackendList] = useState([]);
  const searchParamsRef = useRef<any>();

  useEffect(() => {
    const newParams = queryString.stringify({
      backendID,
      ...(searchParamsRef.current?.query || {}),
      page: searchParamsRef.current?.page,
      pageSize: searchParamsRef.current?.pageSize,
    })
    navigate(`?${newParams}`)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, navigate])

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
      } else {
        message.error(response?.data?.message)
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
        const newParams = {
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        }
        setSearchParams(newParams)
        searchParamsRef.current = newParams;
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
    const newParams = {
      ...searchParams,
      query: {
        backendID: undefined
      }
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getListWorkspace({
      page: 1,
      pageSize: 10,
      query: undefined
    })
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    const newParams = {
      ...searchParams,
      query: values
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getListWorkspace({
      page: 1,
      pageSize: 10,
      query: values,
    })
  }

  useEffect(() => {
    form.setFieldsValue({
      backendID: searchParams?.query?.backendID ? Number(searchParams?.query?.backendID) : undefined
    })
  }, [searchParams?.query, form])

  async function handleSubmit(values, callback) {
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
      callback && callback()
      setOpen(false)
      if (actionType === 'ADD') {
        navigate(`/workspaces/detail/${response?.data?.data?.id}?workspaceName=${response?.data?.data?.name}`)
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
    } else {
      message.error(response?.data?.message)
    }
  }

  function handleChangePage(page, pageSize) {
    getListWorkspace({
      page,
      pageSize,
      query: searchParams?.query
    })
  }

  const pageProps: any = {
    total: searchParams?.total,
    current: searchParams?.page,
    pageSize: searchParams?.pageSize,
    showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} items`,
    showSizeChanger: true,
    pageSizeOptions: [10, 15, 20, 30, 40, 50, 75, 100],
    size: "default",
    style: {
      marginTop: '16px',
      textAlign: 'right'
    },
    onChange: (page, size) => {
      handleChangePage(page, size);
    },
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
        <div style={{ width: '100%', display: 'flex', justifyContent: 'flex-end' }}>
          <Pagination {...pageProps} />
        </div>
      </div>
      <WorkscpaceForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} actionType={actionType} formData={formData} />
    </div>
  )
}

export default Workspaces
