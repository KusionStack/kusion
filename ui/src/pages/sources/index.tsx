import React, { useEffect, useState } from 'react'
import { Button, Form, Input, message, Popconfirm, Space, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import SourceForm from './component/sourceForm'
import { SourceService } from '@kusionstack/kusion-api-client-sdk'

import styles from './styles.module.less'


const SourcesPage = () => {
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: undefined,
  })

  const [dataSource, setDataSource] = useState([])

  async function getResourceList(params) {
    try {
      const response: any = await SourceService.listSource({
        query: {
          sourceName: params?.query?.sourceName,
          page: params?.page || searchParams?.page,
          pageSize: params?.pageSize || searchParams?.pageSize,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.sources);
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
    getResourceList(searchParams)
  }, [])

  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined
    })
    getResourceList({
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
    getResourceList({
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

  function handleChangePage(page: number, pageSize: number) {
    getResourceList({
      ...searchParams,
      page,
      pageSize,
    })
  }

  async function confirmDelete(record) {
    const response: any = await SourceService.deleteSource({
      path: {
        sourceID: record?.id
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getResourceList(searchParams)
    }
  }



  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Url',
      dataIndex: 'remote',
      render: (remoteObj) => {
        return `${remoteObj?.Scheme}//${remoteObj?.Host}${remoteObj?.Path}`
      }
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <Space>
            <Button style={{ padding: '0 5px' }} type='link' onClick={() => handleEdit(record)}>edit</Button>
            <Popconfirm
              title="Delete the source"
              description="Are you sure to delete this source?"
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
      response = await SourceService.updateSource({
        body: {
          id: (formData as any)?.id,
          name: values?.name,
          remote: values?.remote,
          sourceProvider: values?.sourceProvider,
          description: values?.description
        },
        path: {
          sourceID: (formData as any)?.id
        }
      })
    } else {
      response = await SourceService.createSource({
        body: {
          name: values?.name,
          remote: values?.remote,
          sourceProvider: values?.sourceProvider,
          description: values?.description,
        }
      })
    }

    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getResourceList(searchParams)
      setOpen(false)
    } else {
      message.error(response?.data?.messaage)
    }
  }

  function handleCancel() {
    setFormData(undefined)
    setOpen(false)
  }

  const sourceFormProps = {
    open,
    actionType,
    handleSubmit,
    formData,
    handleCancel,
  }

  return (
    <div className={styles.sources}>
      <div className={styles.sources_action}>
        <h3>Sources</h3>
        <div className={styles.sources_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Source
          </Button>
        </div>
      </div>
      <div className={styles.sources_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="sourceName" label="Source Name">
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
      <SourceForm {...sourceFormProps} />
    </div>
  )
}

export default SourcesPage
