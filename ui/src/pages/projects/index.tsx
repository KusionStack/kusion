import React, { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button, Form, Input, message, Space, Table, Popconfirm } from 'antd'
import type { TableColumnsType } from 'antd'
import queryString from 'query-string'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { OrganizationService, ProjectService, SourceService } from '@kusionstack/kusion-api-client-sdk'
import ProjectForm from './components/projectForm'
import DescriptionWithTooltip from '@/components/descriptionWithTooltip'

import styles from "./styles.module.less"



const Projects = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm();
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const { pageSize = 10, page = 1, total = 0, fuzzyName } = queryString.parse(location?.search);
  const [searchParams, setSearchParams] = useState<any>({
    pageSize,
    page,
    query: {
      fuzzyName
    },
    total,
  });
  const [dataSource, setDataSource] = useState([])
  const [organizationList, setOrganizationList] = useState([])
  const [sourceList, setSourceList] = useState([])
  const [open, setOpen] = useState<boolean>(false);
  const searchParamsRef = useRef<any>();

  useEffect(() => {
    const newParams = queryString.stringify({
      fuzzyName,
      ...(searchParamsRef.current?.query || {}),
      page: searchParamsRef.current?.page,
      pageSize: searchParamsRef.current?.pageSize,
    })
    navigate(`?${newParams}`)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, navigate])

  async function createOrganization() {
    const response = await OrganizationService.createOrganization({
      body: {
        name: 'default',
        owners: ['default']
      }
    })
    if (response?.data?.success) {
      getOrganizations()
    } else {
      message.error(response?.data?.message)
    }
  }

  async function getSourceList() {
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

  async function handleSubmit(values, callback) {
    let response: any
    if (actionType === 'EDIT') {
      response = await ProjectService.updateProject({
        body: {
          id: (formData as any)?.id,
          name: values?.name,
          path: values?.path,
          sourceID: values?.projectSource,
          organizationID: organizationList?.[0]?.id,
          description: values?.description,
        },
        path: {
          projectID: (formData as any)?.id
        }
      })
    } else {
      response = await ProjectService.createProject({
        body: {
          name: values?.name,
          path: values?.path,
          sourceID: values?.projectSource,
          organizationID: organizationList?.[0]?.id,
          description: values?.description,
        }
      })
    }
    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getProjectList(searchParams)
      callback && callback()
      setOpen(false)
    } else {
      message.error(response?.data?.message)
    }
  }
  function handleClose() {
    setFormData(undefined)
    setOpen(false)
  }

  function handleReset() {
    form.resetFields();
    const newParams = {
      ...searchParams,
      query: {
        ...searchParams?.query,
        fuzzyName: undefined
      }
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getProjectList({
      page: 1,
      pageSize: 10,
      query: {
        ...searchParams?.query,
        fuzzyName: undefined
      }
    })
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    const newQuery = {
      ...searchParams?.query,
      ...values
    }
    const newParams = {
      ...searchParams,
      query: newQuery
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getProjectList({
      page: 1,
      pageSize: 10,
      query: newQuery
    })
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  function handleAdd() {
    setActionType('ADD')
    setOpen(true)
  }

  function handleEdit(record) {
    setActionType('EDIT')
    setOpen(true)
    setFormData(record)
  }

  function handleChangePage({ current, pageSize }, filters, { field, order }) {
    getProjectList({
      page: current,
      pageSize,
      query: {
        ...searchParams?.query,
        sortBy: field === 'creationTimestamp' ? 'createTimestamp' : field,
        ascending: order === "ascend",
      }
    })
  }

  async function getOrganizations() {
    const response = await OrganizationService.listOrganization()
    if (response?.data?.success) {
      if (response?.data?.data?.organizations?.length > 0) {
        setOrganizationList(response?.data?.data?.organizations)
      } else {
        createOrganization()
      }
    } else {
      message.error(response?.data?.message)
    }
  }

  async function getProjectList(params) {
    try {
      const response: any = await ProjectService.listProject({
        query: {
          ...params?.query,
          pageSize: params?.pageSize || 10,
          page: params?.page,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.projects);
        const newParams = {
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        }
        setSearchParams(newParams)
        searchParamsRef.current = newParams;
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {
    }
  }

  useEffect(() => {
    getOrganizations()
    getSourceList()
    getProjectList(searchParams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    form.setFieldsValue({
      fuzzyName: searchParams?.query?.fuzzyName
    })
  }, [searchParams?.query, form])

  async function confirmDelete(record) {
    const response: any = await ProjectService.deleteProject({
      path: {
        projectID: record?.id,
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getProjectList(searchParams)
    } else {
      message.error(response?.data?.message)
    }
  }


  const columns: TableColumnsType<any> = [
    {
      title: 'Name',
      dataIndex: 'name',
      fixed: 'left',
      sorter: true,
      sortDirections: ['ascend', 'descend', 'ascend'],
      render: (text, record) => {
        return <Button type='link' onClick={() => navigate(`/projects/detail/${record?.id}?projectName=${record?.name}`)}>{text}</Button>
      }
    },
    {
      title: 'Source',
      dataIndex: 'source',
      render: (sourceObj) => {
        const remote = sourceObj?.remote;
        return `${remote?.Scheme}://${remote?.Host}${remote?.Path}`
      }
    },
    {
      title: 'Path',
      dataIndex: 'path',
    },
    {
      title: 'Description',
      dataIndex: 'description',
      render: (desc) => {
        return <DescriptionWithTooltip desc={desc} />
      }
    },
    {
      title: 'Create Time',
      dataIndex: 'creationTimestamp',
      sorter: true,
      sortDirections: ['ascend', 'descend', 'ascend'],
    },
    {
      title: 'Action',
      dataIndex: 'action',
      fixed: 'right',
      width: 150,
      render: (_, record) => {
        return (
          <Space>
            <Button style={{ padding: '0px' }} type='link' onClick={() => handleEdit(record)}>edit</Button>
            <span>/</span>
            <Popconfirm
              title="Delete the project"
              description="Are you sure to delete this project?"
              onConfirm={() => confirmDelete(record)}
              okText="Yes"
              cancelText="No"
            >
              <Button style={{ padding: '0px' }} type='link' danger>delete</Button>
            </Popconfirm>
          </Space>
        )
      },
    }
  ]

  const projectFormProps = {
    open,
    actionType,
    handleSubmit,
    formData,
    handleClose,
    sourceList,
    organizationList,
  }

  function renderTableTitle() {
    const newQuery = {
      fuzzyName: searchParams?.query?.fuzzyName,
    }
    // if (searchParams?.query?.sortBy) {
    //   newQuery[searchParams?.query?.sortBy] = searchParams?.query?.ascending ? 'ascend' : 'descend'
    // }
    const queryList = newQuery && Object.entries(newQuery || {})?.filter(([key, value]) => value !== undefined && value !== null)
    return <div className={styles.projects_content_toolbar}>
      <h4>Project List</h4>
      <div className={styles.projects_content_toolbar_list}>
        {
          queryList?.map(([key, value]) => {
            return <div className={styles.projects_content_toolbar_item}>
              {key === 'fuzzyName' ? 'name' : key}: {value as string}
              <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} /></div>
          })
        }
      </div>
      {
        queryList?.length > 0 && (
          <div className={styles.projects_content_toolbar_clear}>
            <Button type='link' onClick={handleReset} style={{ paddingLeft: 0 }}>Clear</Button>
          </div>
        )
      }
    </div>
  }

  return (
    <div className={styles.projects}>
      <div className={styles.projects_action}>
        <h3>Projects</h3>
        <div className={styles.projects_action_create}>
          <Button type='primary' onClick={handleAdd}>
            <PlusOutlined /> New Project
          </Button>
        </div>
      </div>
      <div className={styles.projects_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="fuzzyName" label="Project Name">
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
        <Table
          rowKey="id"
          title={renderTableTitle}
          columns={columns}
          dataSource={dataSource}
          scroll={{ x: 1300 }}
          onChange={handleChangePage}
          pagination={{
            total: Number(searchParams?.total),
            current: Number(searchParams?.page),
            pageSize: Number(searchParams?.pageSize),
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} items`,
            showSizeChanger: true,
            pageSizeOptions: [10, 15, 20, 30, 40, 50, 75, 100],
            size: "default",
            style: {
              marginRight: 16,
              textAlign: 'right'
            },
          }}
        />
      </div>
      <ProjectForm {...projectFormProps} />
    </div>
  )
}

export default Projects
