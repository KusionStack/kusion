import React, { useEffect, useState } from 'react'
import { Button, Card, Col, Input, Popconfirm, Row, Tooltip } from 'antd'
import {
  DeleteOutlined,
  PlusOutlined,
  SortDescendingOutlined,
  SortAscendingOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';

import styles from './styles.module.less'
import WorkspaceCard from './components/workspaceCard';
import WorkscpaceForm from './components/workscpaceForm';

const orderIconStyle: React.CSSProperties = {
  marginLeft: 0,
}

const Workspaces = () => {
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState<string>('')

  const [sortParams, setSortParams] = useState<any>({
    orderBy: 'name',
    isAsc: true,
  })


  async function getListWorkspace() {
    try {
      const sources = await WorkspaceService.listWorkspace();
      console.log('WorkspaceService:', sources.data);
    } catch (error) {
      console.error('Error:', error);
    }
  }

  useEffect(() => {
    getListWorkspace()
  }, [])

  function handleAdd() {
    setOpen(true)
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


  function conversionArray(baseArray, n) {
    const len = baseArray.length
    const lineNum = len % n === 0 ? len / n : Math.floor(len / n + 1)
    const res = []
    for (let i = 0; i < lineNum; i++) {
      const temp = baseArray.slice(i * n, i * n + n)
      res.push(temp)
    }
    return res
  }

  function handleChange(event) {
    setKeyword(event?.target.value)
  }

  function handleSubmit(values) {
    console.log(values, "=====handleSubmit values=====")
  }
  function handleClose() {
    setOpen(false)
  }

  const arrayColByN = conversionArray([1, 2, 3, 4], 4)
  console.log(arrayColByN, '====arrayColByN====')

  const mockDesc =
    '这是一段描述文字超长文本测试测试测试测试测试测试测试测试测试测试测试测试测试测试测试测试'

  return (
    <div className={styles.kusion_workspace_container}>
      <div className={styles.kusion_workspace_toolbar}>
        <Button type="primary" onClick={handleAdd}>
          <PlusOutlined /> New Workspace
        </Button>
        <div className={styles.kusion_workspace_toolbar_right}>
          <Input
            placeholder={'关键字搜索'}
            suffix={<SearchOutlined />}
            style={{ width: 260 }}
            value={keyword}
            onChange={handleChange}
            allowClear
          />
          <div className={styles.kusion_action_bar_right_sort}>{renderDateSort()}</div>
        </div>
      </div>
      <div className={styles.kusion_workspace_content}>
        {arrayColByN?.map((item, index) => {
          return (
            <Row
              key={index}
              gutter={{ xs: 8, sm: 16, md: 24, lg: 32 }}
              style={{ marginBottom: 20 }}
            >
              {item?.map((innerItem, innerIndex) => {
                return (
                  <Col key={innerIndex} className="gutter-row" span={6}>
                    <WorkspaceCard title="title" desc={mockDesc} createDate="20241218" nickName="测试" />
                  </Col>
                )
              })}
            </Row>
          )
        })}
      </div>
      <WorkscpaceForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} />
    </div>
  )
}

export default Workspaces
