import React, { useState } from 'react'
// import { AutoComplete, Input, message, Space, Tag } from 'antd'
// import {
//   DoubleLeftOutlined,
//   DoubleRightOutlined,
//   CloseOutlined,
// } from '@ant-design/icons'

import styles from './styles.module.less'
import { Button, Input, Select, Space, Table } from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import ModuleForm from './component/moduleForm'

const { Option } = Select

const ModulePanelContent = () => {
  const [keyword, setKeyword] = useState<string>('')

  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()

  function handleChange(event) {
    setKeyword(event?.target.value)
  }

  function handleAdd() {
    console.log('新增Source')
    setActionType('ADD')
    setOpen(true)
  }
  function handleEdit(record) {
    console.log(record, '编辑')
    setActionType('EDIT')
    setOpen(true)
    setFormData(record)
  }
  function handleDetail(record) {
    console.log(record, '查看详情')
    setActionType('CHECK')
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
          <Space>
            <a onClick={() => handleDetail(record)}>详情</a>
            <a onClick={() => handleEdit(record)}>编辑</a>
            <a
              href={record?.documentUrl || 'https://www.alipay.com'}
              target="_blank"
              rel="noreferrer"
            >
              文档
            </a>
          </Space>
        )
      },
    },
  ]

  const dataSource = [
    {
      name: 'Module1',
      registry: 'https://www.alipay.com',
      publishTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Module2',
      registry: 'https://www.alipay.com',
      publishTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Module3',
      registry: 'https://www.alipay.com',
      publishTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Module4',
      registry: 'https://www.alipay.com',
      publishTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Module5',
      registry: 'https://www.alipay.com',
      publishTime: new Date().toLocaleDateString(),
    },
  ]

  function handleSubmit(values) {
    console.log(values, 'handleSubmit')
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
    <div className={styles.panel_content}>
      <div className={styles.tool_bar}>
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
              <Select style={{ width: 200 }} allowClear>
                <Option key={'ts1'} value={'ts1'}>
                  测试1
                </Option>
                <Option key={'ts2'} value={'ts2'}>
                  测试2
                </Option>
              </Select>
            </Space>
          </div>
        </div>
        <div className={styles.right}>
          <div className={styles.tool_bar_add}>
            <Button type="primary" onClick={handleAdd}>
              <PlusOutlined /> 新增Module
            </Button>
          </div>
        </div>
      </div>
      <Table columns={colums} dataSource={dataSource} />
      <ModuleForm {...sourceFormProps} />
    </div>
  )
}

export default ModulePanelContent
