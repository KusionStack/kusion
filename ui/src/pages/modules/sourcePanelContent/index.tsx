import React, { useState } from 'react'
// import { AutoComplete, Input, message, Space, Tag } from 'antd'
// import {
//   DoubleLeftOutlined,
//   DoubleRightOutlined,
//   CloseOutlined,
// } from '@ant-design/icons'

import styles from './styles.module.less'
import { Button, Input, Space, Table } from 'antd'
import {
  SearchOutlined,
  SortDescendingOutlined,
  SortAscendingOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import SourceForm from './component/sourceForm'

const orderIconStyle: React.CSSProperties = {
  marginLeft: 0,
}

const SourcesPanelContent = () => {
  const [keyword, setKeyword] = useState<string>('')
  const [sortParams, setSortParams] = useState<any>({
    orderBy: 'name',
    isAsc: true,
  })
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()

  function handleChange(event) {
    setKeyword(event?.target.value)
  }

  function handleSort(key: any) {
    setSortParams({
      orderBy: key,
      isAsc: !sortParams?.isAsc,
    })
    // const groups = allResourcesData?.groups?.sort((a, b) =>
    //   sortParams?.isAsc
    //     ? b.title.localeCompare(a.title)
    //     : a.title.localeCompare(b.title),
    // )
    // setAllResourcesData({
    //   fields: allResourcesData?.fields,
    //   groups,
    // })
  }

  function renderSort() {
    return (
      <Button
        type="link"
        style={{ color: '#646566', marginRight: 0 }}
        onClick={() => handleSort('name')}
      >
        名称
        {sortParams?.orderBy === 'name' &&
          (sortParams?.isAsc ? (
            <SortDescendingOutlined style={orderIconStyle} />
          ) : (
            <SortAscendingOutlined style={orderIconStyle} />
          ))}
      </Button>
    )
  }

  function renderDateSort() {
    return (
      <Button
        type="link"
        style={{ color: '#646566', marginRight: 0 }}
        onClick={() => handleSort('date')}
      >
        日期
        {sortParams?.orderBy === 'date' &&
          (sortParams?.isAsc ? (
            <SortDescendingOutlined style={orderIconStyle} />
          ) : (
            <SortAscendingOutlined style={orderIconStyle} />
          ))}
      </Button>
    )
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
      title: 'Url',
      dataIndex: 'url',
    },
    {
      title: 'Modify Time',
      dataIndex: 'modifyTime',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => {
        return (
          <Space>
            <a onClick={() => handleDetail(record)}>详情</a>
            <a onClick={() => handleEdit(record)}>编辑</a>
          </Space>
        )
      },
    },
  ]

  const dataSource = [
    {
      name: 'Source1',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source2',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source3',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source4',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source5',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source6',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source7',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source8',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source9',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
    },
    {
      name: 'Source10',
      url: 'https://www.alipay.com',
      modifyTime: new Date().toLocaleDateString(),
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
            <Input
              placeholder={'关键字搜索'}
              suffix={<SearchOutlined />}
              style={{ width: 260 }}
              value={keyword}
              onChange={handleChange}
              allowClear
            />
          </div>
          <div className={styles.tool_bar_sort}>
            <div className={styles.action_bar_right_sort}>{renderSort()}</div>
            <div className={styles.action_bar_right_sort}>
              {renderDateSort()}
            </div>
          </div>
        </div>
        <div className={styles.right}>
          <div className={styles.tool_bar_add}>
            <Button type="primary" onClick={handleAdd}>
              <PlusOutlined /> 新增Source
            </Button>
          </div>
        </div>
      </div>
      <Table columns={colums} dataSource={dataSource} />
      <SourceForm {...sourceFormProps} />
    </div>
  )
}

export default SourcesPanelContent
