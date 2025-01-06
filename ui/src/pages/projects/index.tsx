import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, message, Space, Table, Select } from 'antd'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { OrganizationService, ProjectService, SourceService } from '@kusionstack/kusion-api-client-sdk'
import ProjectForm from './components/projectForm'

import styles from "./styles.module.less"

const Projects = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [searchParams, setSearchParams] = useState({
    pageSize: 10,
    page: 1,
    query: {},
    total: undefined,
  });
  const [dataSource, setDataSource] = useState([])
  const [organizationList, setOrganizationList] = useState([])
  const [sourceList, setSourceList] = useState([])
  const [open, setOpen] = useState<boolean>(false);

  async function getResourceList() {
    try {
      const response: any = await SourceService.listSource({
        ...searchParams,
        query: {
          page: 1,
          pageSize: 10000
        }
      });
      if (response?.data?.success) {
        setSourceList(response?.data?.data?.sources);
      } else {
        message.error(response?.data?.messaage)
      }
    } catch (error) {

    }
  }

  async function handleSubmit(values) {
    console.log(values, "handleSubmit")
    const response: any = await ProjectService.createProject({
      body: {
        domain: values?.name,
      }
    })
    if (response?.data?.success) {
      message.success('Create Successful')
      getProjectList(searchParams)
      setOpen(false)
    } else {
      message.error(response?.data?.message)
    }
  }
  function handleClose() {
    setOpen(false)
  }
  function handleCreate() {
    setOpen(true)
  }
  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined
    })
    getProjectList({
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
    getProjectList({
      page: 1,
      pageSize: 20,
      query: values,
    })
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  async function getOrganizations() {
    const response = await OrganizationService.listOrganization()
    if (response?.data?.success) {
      setOrganizationList(response?.data?.data?.organizations)
    } else {
      message.error(response?.data?.message)
    }
  }

  async function getProjectList(params) {
    try {
      const response: any = await ProjectService.listProject({
        query: {
          ...params?.query,
          pageSize: params?.pageSize || 20,
          page: params?.page,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.projects);
        setSearchParams({
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        })
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {
    }
  }

  useEffect(() => {
    getOrganizations()
    getResourceList()
    getProjectList(searchParams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])


  const colums = [
    {
      title: 'Name',
      dataIndex: 'name',
      render: (text, record) => {
        return <Button type='link' onClick={() => navigate(`/projects/detail/${record?.id}?projectName=${record?.name}`)}>{text}</Button>
      }
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Organization',
      dataIndex: 'organization',
      render: (organization) => {
        return <div>{organization?.name}</div>
      }
    },
    {
      title: 'Owners',
      dataIndex: 'owners',
    },
    {
      title: 'Create Time',
      dataIndex: 'creationTimestamp',
    },
  ]

  function renderTableTitle(currentPageData) {
    const queryList = searchParams && Object.entries(searchParams?.query || {})?.filter(([key, value]) => value)
    return <div className={styles.projects_content_toolbar}>
      <h4>Project List</h4>
      <div className={styles.projects_content_toolbar_list}>
        {
          queryList?.map(([key, value]) => {
            return <div className={styles.projects_content_toolbar_item}>
              {key}: {value as string}
              <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} /></div>
          })
        }
      </div>
      {
        queryList?.length > 0 && <div className={styles.projects_content_toolbar_clear}>
          <Button type='link' onClick={handleReset} style={{ paddingLeft: 0 }}>Clear</Button>
        </div>
      }
    </div>
  }

  return (
    <div className={styles.projects}>
      <div className={styles.projects_action}>
        <h3>Projects</h3>
        <div className={styles.projects_action_create}>
          <Button type='primary' onClick={handleCreate}><PlusOutlined /> Create Projects</Button>
        </div>
      </div>
      <div className={styles.projects_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="name" label="Project Name">
              <Input />
            </Form.Item>
            <Form.Item name="organization" label="Organization">
              <Select style={{ width: 150 }}>
                {
                  organizationList?.map(item => <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item name="owner" label="Owner">
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
      <div className={styles.projects_content}>
        <Table rowKey="id" title={renderTableTitle} columns={colums} dataSource={dataSource} />
      </div>
      <ProjectForm open={open}
        sourceList={sourceList}
        organizationList={organizationList}
        handleSubmit={handleSubmit}
        handleClose={handleClose}
      />
    </div>
  )
}

export default Projects
