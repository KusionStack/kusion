import React, { useEffect, useState } from 'react'

import { Button, DatePicker, Form, Input, message, Space, Table, } from 'antd'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { StackService } from '@kusionstack/kusion-api-client-sdk'

import styles from "./styles.module.less"

const Runs = () => {
  const [form] = Form.useForm();
  const [dataSource, setDataSource] = useState([])
  const [open, setOpen] = useState<boolean>(false);
  const [searchParams, setSearchParams] = useState({
    pageSize: 20,
    page: 1,
    query: undefined,
    total: 0,
  })

  function handleSubmit(values) {
    console.log(values, "handleSubmit")
  }
  function handleClose() {
    console.log("handleClose")
    setOpen(false)
  }
  function handleCreate() {
    setOpen(true)
    console.log("=====handleCreate=====")
  }
  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined,
    })
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    setSearchParams({
      ...searchParams,
      query: values,
    })
    console.log(values, "=====handleSearch=====")
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  async function getListRun(params) {
    console.log(params, "=====params=====")
    try {
      const response: any = await StackService.listRun({
        ...params?.query,
        pageSize: params?.pageSize,
        currentPage: params?.page,
      });
      if (response?.data?.success) {
        console.log(response?.data?.data, "======StackService resData====")
        setDataSource(response?.data?.data?.runs);
        setSearchParams({
          ...searchParams,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        })
      } else {
        message.error('请求失败')
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    getListRun({})
  }, [])

  function handleChangePage(page: number, pageSize: number) {
    getListRun({
      ...searchParams,
      page,
      pageSize,
    })
  }

  function handleCheckDetail(record) {
    console.log(record, "-=======record========")
  }


  const colums = [
    {
      title: 'Runs ID',
      dataIndex: 'id',
    },
    {
      title: 'Type',
      dataIndex: 'type',
    },
    {
      title: 'Create Time',
      dataIndex: 'creationTimestamp',
    },
    {
      title: 'Status',
      dataIndex: 'status',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => <Button type='link' onClick={() => handleCheckDetail(record)}>Detail</Button>
    },
  ]

  function handleCreateRuns() {
    console.log("========handleCreateRuns=========")
  }

  function renderTableTitle() {
    return <div className={styles.project_runs_toolbar}>
      <div className={styles.project_runs_toolbar_left}>
        <div className={styles.project_runs_result}>共找到<Button style={{ padding: 4 }} type='link'>{searchParams?.total}</Button>个结果</div>
        <div className={styles.projects_content_toolbar_list}>
          {
            searchParams?.query && Object.entries(searchParams?.query)?.filter(([key, _value]) => _value)?.map(([key, __value]: any) => {
              return (
                <div className={styles.projects_content_toolbar_item}>
                  {key}: {__value}
                  <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} />
                </div>
              )
            })
          }
        </div>
        {
          searchParams?.query && <div className={styles.projects_content_toolbar_clear}>
            <Button type='link' onClick={handleReset} style={{ paddingLeft: 0 }}>Clear</Button>
          </div>
        }
      </div>
      <div className={styles.projects_content_toolbar_create}>
        <Button onClick={handleCreateRuns}><PlusOutlined /> Create Runs</Button>
      </div>
    </div>
  }



  return (
    <div className={styles.project_runs}>
      {/* Search Form block*/}
      <div className={styles.project_runs_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="type" label="Type">
              <Input />
            </Form.Item>
            <Form.Item name="status" label="Status">
              <Input />
            </Form.Item>
            <Form.Item name="createTime" label="Create Time">
              <DatePicker.RangePicker />
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
      {/* Content block */}
      <div className={styles.project_runs_content}>
        {renderTableTitle()}
        <Table
          size='small'
          columns={colums}
          dataSource={dataSource}
          pagination={
            {
              style: { paddingRight: 20 },
              total: searchParams?.total,
              showTotal: (total: number, range: any[]) => `${range?.[0]}-${range?.[1]} Total ${total} `,
              pageSize: searchParams?.pageSize,
              current: searchParams?.page,
              onChange: handleChangePage,
            }
          }
        />
      </div>
    </div>
  )
}

export default Runs
