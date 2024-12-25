import React, { useEffect, useState } from 'react'
import { Button, Card, Input, message, Space, Table } from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import { ModuleService } from '@kusionstack/kusion-api-client-sdk'
import ModuleForm from './component/moduleForm'
import { debounce } from "lodash"

import styles from './styles.module.less'


const ModulePage = () => {
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

  const [moduleList, setModuleList] = useState([])

  async function getModuleList(params) {
    try {
      const response: any = await ModuleService.listModule();
      if (response?.data?.success) {
        setModuleList(response?.data?.data);
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

  const handleChange = debounce((event) => {
    const val = event?.target.value;
    setKeyword(val)
    getModuleList({
      ...searchParams,
      query: {
        keyword: val
      }
    })
  }, 200)

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
      title: 'Registry',
      dataIndex: 'registry',
      render: (_, record) => {
        return record?.url?.Path;
      }
    },
    {
      title: 'Publish Time',
      dataIndex: 'publishTime',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <>
            <Button type='link' onClick={() => handleEdit(record)}>编辑</Button>
            <Button type='link' href={record?.doc?.Path} target='_blank'>文档</Button>
          </>
        )
      },
    },
  ]



  async function handleSubmit(values) {
    console.log(values, 'Sources handleSubmit')
    const response: any = await ModuleService.createModule({
      body: values
    })
    if (response?.data?.success) {
      message.success('Create Success')
      getModuleList({})
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
      <div className={styles.modules_container}>
        <div className={styles.modules_toolbar}>
          <div className={styles.left}>
            <div className={styles.tool_bar_search}>
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
          <div className={styles.right}>
            <div className={styles.tool_bar_add}>
              <Button type="primary" onClick={handleAdd}>
                <PlusOutlined /> New Module
              </Button>
            </div>
          </div>
        </div>
        <Table columns={colums} dataSource={moduleList} />
        <ModuleForm {...sourceFormProps} />
      </div>
    </Card>
  )
}

export default ModulePage
