import React, { useEffect, useState } from 'react'
import { Button, Form, Input, message, Popconfirm, Space, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { ModuleService } from '@kusionstack/kusion-api-client-sdk'
import ModuleForm from './component/moduleForm'

import styles from './styles.module.less'


const ModulePage = () => {
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [searchParams, setSearchParams] = useState({
    pageSize: 10,
    page: 1,
    query: undefined,
    total: undefined,
  })

  const [dataSource, setDataSource] = useState([])

  async function getModuleList(params) {
    try {
      const response: any = await ModuleService.listModule({
        query: {
          moduleName: params?.query?.moduleName,
          page: params?.page || searchParams?.page,
          pageSize: params?.pageSize || searchParams?.pageSize,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.modules);
        setSearchParams({
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        })
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    getModuleList({})
  }, [])

  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined
    })
    getModuleList({
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
    getModuleList({
      page: 1,
      pageSize: 10,
      query: values,
    })
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

  async function confirmDelete(record) {
    const response: any = await ModuleService.deleteModule({
      path: {
        moduleName: record?.name
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getModuleList(searchParams)
    }
  }



  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
    },
    {
      title: 'Registry',
      dataIndex: 'registry',
      render: (_, record) => {
        return `${record?.url?.Scheme}://${record?.url?.Host}${record?.url?.Path}`;
      }
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <Space>
            <Button type='link' onClick={() => handleEdit(record)}>edit</Button>
            <Button type='link' href={record?.doc?.Path} target='_blank'>doc</Button>
            <Popconfirm
              title="Delete the module"
              description="Are you sure to delete this module?"
              onConfirm={() => confirmDelete(record)}
              okText="Yes"
              cancelText="No"
            >
              <Button type='link' danger>Delete</Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]



  async function handleSubmit(values) {
    let response: any
    if (actionType === 'EDIT') {
      response = await ModuleService.updateModule({
        body: values,
        path: {
          moduleName: (formData as any)?.name
        }
      })
    } else {
      response = await ModuleService.createModule({
        body: values,
      })
    }

    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getModuleList({})
      setOpen(false)
    } else {
      message.error(response?.data?.messaage)
    }
  }

  function handleCancel() {
    setFormData(undefined)
    setOpen(false)
  }

  function handleChangePage(page: number, pageSize: number) {
    getModuleList({
      ...searchParams,
      page,
      pageSize,
    })
  }

  const sourceFormProps = {
    open,
    actionType,
    handleSubmit,
    formData,
    handleCancel,
  }

  return (
    <div className={styles.modules}>
      <div className={styles.modules_action}>
        <h3>Modules</h3>
        <div className={styles.modules_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Module
          </Button>
        </div>
      </div>
      <div className={styles.modules_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="moduleName" label="Module Name">
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
      <div className={styles.modules_content}>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={dataSource}
          pagination={
            {
              style: { paddingRight: 20 },
              total: searchParams?.total,
              showTotal: (total: number, range: any[]) => `${range?.[0]}-${range?.[1]} Total ${total} `,
              pageSize: searchParams?.pageSize,
              current: searchParams?.page,
              onChange: handleChangePage,
            }
          }
        />
      </div>
      <ModuleForm {...sourceFormProps} />
    </div>
  )
}

export default ModulePage
