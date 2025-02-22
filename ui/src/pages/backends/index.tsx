import React, { useEffect, useState } from 'react'
import { Button, message, Popconfirm, Space, Table, Tag } from 'antd'
import type { TableColumnsType } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { BackendService } from '@kusionstack/kusion-api-client-sdk'
import { josn2yaml } from '@/utils/tools'
import DescriptionWithTooltip from '@/components/descriptionWithTooltip'
import BackendForm from './component/backendForm'
import ConfigYamlDrawer from './component/configYamlDrawer'

import styles from './styles.module.less'

const BackendsPage = () => {
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
        query: {
          sourceName: params?.query?.sourceName,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.backends);
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
    getBackendList(searchParams)
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

  function handleShowConfig(record) {
    setCurrentRecord(record)
    setConfigOpen(true)
  }

  async function confirmDelete(id) {
    const response: any = await BackendService.deleteBackend({
      path: {
        backendID: id
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getBackendList(searchParams)
    } else {
      message.error(response?.data?.message)
    }
  }


  const columns: TableColumnsType<any> = [
    {
      title: 'Name',
      dataIndex: 'name',
      fixed: 'left',
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
      render: (desc) => {
        return <DescriptionWithTooltip desc={desc} width={500} />
      }
    },
    {
      title: 'Config',
      dataIndex: 'config',
      render: (_, record) => {
        return <Button style={{ padding: '0px' }} type='link' onClick={() => handleShowConfig(record)}>Detail</Button>
      }
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
              title="Delete the backend"
              description="Are you sure to delete this backend?"
              onConfirm={() => confirmDelete(record?.id)}
              okText="Yes"
              cancelText="No"
            >
              <Button style={{ padding: '0px' }} type='link' danger>delete</Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  async function handleSubmit(values, callback) {
    let response: any
    let bodyParams: any = {};
    try {
      bodyParams = {
        name: values?.name,
        backendConfig: {
          configs: values?.configs,
          type: values?.type,
        },
        description: values?.description
      }
    } catch (error) {
      console.log(error)
    }
    if (actionType === 'EDIT') {
      response = await BackendService.updateBackend({
        body: {
          id: (formData as any)?.id,
          ...bodyParams,
        },
        path: {
          backendID: (formData as any)?.id
        }
      })
    } else {
      response = await BackendService.createBackend({
        body: bodyParams
      })
    }

    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getBackendList(searchParams)
      setOpen(false)
      callback && callback()
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
    yamlData: (currentRecord as any)?.backendConfig?.configs ? josn2yaml(JSON.stringify((currentRecord as any)?.backendConfig?.configs))?.data : ''
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
      <div className={styles.modules_content}>
        <Table
          title={() => <h4>Backend List</h4>}
          rowKey="id"
          columns={columns}
          dataSource={dataSource}
          scroll={{ x: 1300 }}
        />
      </div>
      <ConfigYamlDrawer {...configYamlProps} />
      <BackendForm {...sourceFormProps} />
    </div>
  )
}

export default BackendsPage
