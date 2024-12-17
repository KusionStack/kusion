import React, { useEffect, useState } from 'react'
import { Button, Card, Col, Input, Popconfirm, Row, Tooltip } from 'antd'
import {
  DeleteOutlined,
  PlusOutlined,
  SortDescendingOutlined,
  SortAscendingOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { SourceService, ResourceService } from '@kusionstack/kusion-api-client-sdk';

import styles from './styles.module.less'
import G6Tree from '@/components/g6Tree'

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

  async function getSources() {
    try {
      const sources = await SourceService.listSource();
      console.log('Sources:', sources.data);
    } catch (error) {
      console.error('Error:', error);
    }
  }
  
  async function getResourceGraph() {
    try {
      const sources = await ResourceService.getResourceGraph({query: '1'} as any);
      console.log('ResourceService:', sources.data);
    } catch (error) {
      console.error('Error:', error);
    }
  }

  useEffect(() => {
    getSources()
    getResourceGraph()
  }, [])

  function handleAdd() {
    setOpen(true)
  }

  function confirmDelete() {
    console.log('删除成功')
  }

  function cancel() {
    console.log('取消操作')
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

  const arrayColByN = conversionArray([1, 2, 3, 4], 4)
  console.log(arrayColByN, '====arrayColByN====')

  const mockDesc =
    '这是一段描述文字超长文本测试测试测试测试测试测试测试测试测试测试测试测试测试测试测试测试'

  return (
    <>
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
                    <Card
                      title={`Workspace ${innerItem}`}
                      extra={
                        <Popconfirm
                          title="Delete the workspace"
                          description="Are you sure to delete this workspace?"
                          onConfirm={confirmDelete}
                          onCancel={cancel}
                          okText="Yes"
                          cancelText="No"
                        >
                          <DeleteOutlined />
                        </Popconfirm>
                      }
                    >
                      <Tooltip title={mockDesc}>
                        <div className={styles.kusion_card_content_desc}>
                          {mockDesc}
                        </div>
                      </Tooltip>
                    </Card>
                  </Col>
                )
              })}
            </Row>
          )
        })}
      </div>
      <G6Tree />
    </>
  )
}

export default Workspaces
