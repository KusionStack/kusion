import React, { useEffect, useState } from 'react'
import { Button, Form, Input, message, Space, Table, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { BackendService } from '@kusionstack/kusion-api-client-sdk'
import BackendForm from './component/backendForm'
import ConfigYamlDrawer from './component/configYamlDrawer'

import styles from './styles.module.less'

const BackendsPage = () => {
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
  const [currentRecord, setCurrentRecord] = useState({})
  const [configOpen, setConfigOpen] = useState(false)

  async function getBackendList(params) {
    try {
      const response: any = await BackendService.listBackend({
        ...searchParams,
        query: {
          sourceName: params?.query?.sourceName,
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
    getBackendList(searchParams)
  }, [])

  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined
    })
    getBackendList({
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
    getBackendList({
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

  function handleShowConfig(record) {
    setCurrentRecord(record)
    setConfigOpen(true)
  }


  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
    },
    {
      title: 'Type',
      dataIndex: 'type',
      render: (_, record) => {
        return <Tag>{record?.backendConfig?.type}</Tag>
      }
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Config',
      dataIndex: 'config',
      render: (_, record) => {
        return <Button type='link' onClick={() => handleShowConfig(record)}>Detail</Button>
      }
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <Button style={{ padding: '0 5px' }} type='link' onClick={() => handleEdit(record)}>edit</Button>
        )
      },
    },
  ]

  async function handleSubmit(values) {
    let response: any
    if (actionType === 'EDIT') {
      response = await BackendService.updateBackend({
        body: {
          id: (formData as any)?.id,
          name: values?.name,
          backendConfig: {
            configs: values?.configs,
            type: values?.type,
          },
          description: values?.description
        },
        path: {
          backendID: (formData as any)?.id
        }
      })
    } else {
      response = await BackendService.createBackend({
        body: {
          name: values?.name,
          backendConfig: {
            configs: values?.configs,
            type: values?.type,
          },
          description: values?.description,
        }
      })
    }

    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getBackendList(searchParams)
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

  const configYamlProps = {
    open: configOpen,
    handleClose: () => setConfigOpen(false),
    yamlData: (currentRecord as any)?.backendConfig?.configs
  }

  return (
    <div className={styles.sources}>
      <div className={styles.sources_action}>
        <h3>Backends</h3>
        <div className={styles.sources_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Backend
          </Button>
        </div>
      </div>
      <div className={styles.sources_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="backendName" label="Backend Name">
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
        <Table columns={columns} dataSource={dataSource} />
      </div>
      <ConfigYamlDrawer {...configYamlProps} />
      <BackendForm {...sourceFormProps} />
    </div>
  )
}

export default BackendsPage
