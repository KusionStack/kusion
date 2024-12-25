import React, { useEffect, useState } from 'react'
// import { AutoComplete, Input, message, Space, Tag } from 'antd'
// import {
//   DoubleLeftOutlined,
//   DoubleRightOutlined,
//   CloseOutlined,
// } from '@ant-design/icons'

import styles from './styles.module.less'
import { Button, Card, Input, message, Select, Space, Table } from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import SourceForm from './component/sourceForm'
import { SourceService } from '@kusionstack/kusion-api-client-sdk'

const { Option } = Select

const SourcesPage = () => {
  const [keyword, setKeyword] = useState<string>('')

  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: 0,
  })

  const [resourceList, setResourceList] = useState([])

  async function getResourceList(params) {
    try {
      const response: any = await SourceService.listSource();
      if (response?.data?.success) {
        setResourceList(response?.data?.data);
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
    getResourceList({})
  }, [])

  function handleChange(event) {
    setKeyword(event?.target.value)
  }

  function handleAdd() {
    setActionType('ADD')
    setOpen(true)
  }
  function handleEdit(record) {
    console.log(record, '编辑')
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
    console.log(values, 'Sources handleSubmit')
    const response: any = await SourceService.createSource({
      body: {
        name: values?.name,
        remote: values?.remote,
        sourceProvider: values?.sourceProvider
      }
    })
    if (response?.data?.success) {
      message.success('Create Success')
      getResourceList({})
      setOpen(false)
    } else {
      message.error(response?.data?.messaage || '请求失败')
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
            <div className={styles.tool_bar_add}>
              <Button type="primary" onClick={handleAdd}>
                <PlusOutlined /> New Source
              </Button>
            </div>
          </div>
          <div className={styles.sources_toolbar_right}>
            <Space>
              <Input
                placeholder={'关键字搜索'}
                suffix={<SearchOutlined />}
                style={{ width: 260 }}
                value={keyword}
                onChange={handleChange}
                allowClear
              />
            </Space>
          </div>
        </div>
        <Table size='small' columns={colums} dataSource={resourceList} />
        <SourceForm {...sourceFormProps} />
      </div>
    </Card>
  )
}

export default SourcesPage
