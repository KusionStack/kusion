import React, { useEffect, useState } from 'react'
import { Button, Card, Input, message, Space, Table } from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import { ModuleService } from '@kusionstack/kusion-api-client-sdk'
import ModuleForm from './component/moduleForm'
import { debounce } from "lodash"

import styles from './styles.module.less'


const ModulePage = () => {
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: undefined,
  })

  const [moduleList, setModuleList] = useState([])

  async function getModuleList(params) {
    try {
      const response: any = await ModuleService.listModule({
        query: {
          moduleName: params?.query?.moduleName
        }
      });
      if (response?.data?.success) {
        setModuleList(response?.data?.data?.modules);
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
    setSearchParams({
      ...searchParams,
      query: {
        moduleName: val
      }
    })
    getModuleList({
      ...searchParams,
      query: {
        moduleName: val
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



  const columns = [
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
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <>
            <Button type='link' onClick={() => handleEdit(record)}>edit</Button>
            <Button type='link' href={record?.doc?.Path} target='_blank'>doc</Button>
          </>
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
                  placeholder='keyword search'
                  suffix={<SearchOutlined />}
                  style={{ width: 260 }}
                  value={searchParams?.query?.moduleName}
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
        <Table columns={columns} dataSource={moduleList} />
        <ModuleForm {...sourceFormProps} />
      </div>
    </Card>
  )
}

export default ModulePage
