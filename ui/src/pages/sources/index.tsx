import React, { useEffect, useState } from 'react'
import { Button, Card, Input, message, Space, Table } from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import SourceForm from './component/sourceForm'
import { SourceService } from '@kusionstack/kusion-api-client-sdk'
import { debounce } from "lodash"

import styles from './styles.module.less'


const SourcesPage = () => {
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: undefined,
  })

  const [resourceList, setResourceList] = useState([])

  async function getResourceList(params) {
    try {
      const response: any = await SourceService.listSource({
        ...searchParams,
        query: {
          sourceName: searchParams?.query?.sourceName,
        }
      });
      if (response?.data?.success) {
        setResourceList(response?.data?.data?.sources);
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

  const handleChange = debounce((event) => {
    const val = event?.target.value;
    setSearchParams({
      ...searchParams,
      query: {
        sourceName: val
      }
    })
    getResourceList({
      ...searchParams,
      query: {
        sourceName: val
      }
    })
  }, 800)

  function handleAdd() {
    setActionType('ADD')
    setOpen(true)
  }
  function handleEdit(record) {
    setActionType('EDIT')
    setOpen(true)
    setFormData(record)
  }



  const colums = [
    {
      title: 'Name',
      dataIndex: 'name',
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Creation Time',
      dataIndex: 'creationTimestamp',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <Button style={{ padding: '0 5px' }} type='link' onClick={() => handleEdit(record)}>编辑</Button>
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
          id: (formData as any)?.id
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
    <Card>
      <div className={styles.sources_container}>
        <div className={styles.sources_toolbar}>
          <div className={styles.sources_toolbar_left}>
            <Space>
              <Input
                placeholder={'关键字搜索'}
                suffix={<SearchOutlined />}
                style={{ width: 260 }}
                value={searchParams?.query?.sourceName}
                onChange={handleChange}
                allowClear
              />
            </Space>
          </div>
          <div className={styles.sources_toolbar_right}>
            <div className={styles.tool_bar_add}>
              <Button type="primary" onClick={handleAdd}>
                <PlusOutlined /> New Source
              </Button>
            </div>
          </div>
        </div>
        <Table size='small' columns={colums} dataSource={resourceList} />
        <SourceForm {...sourceFormProps} />
      </div>
    </Card>
  )
}

export default SourcesPage
