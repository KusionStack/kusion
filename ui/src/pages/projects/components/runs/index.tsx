import React, { useEffect, useState } from 'react'
import { Button, DatePicker, Form, message, Select, Space, Table, Tag, } from 'antd'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { StackService } from '@kusionstack/kusion-api-client-sdk'
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { RUNS_STATUS_MAP, RUNS_TYPES } from '@/utils/constants'
import GenerateDetail from './generateDetail'
import PreviewDetail from './previewDetail'
import RunsForm from './runsForm'

import styles from "./styles.module.less"

// use plugin
dayjs.extend(utc);
dayjs.extend(timezone);

const timeFormatter = 'YYYY-MM-DDTHH:mm:ssZ'

const Runs = ({ stackId }) => {
  const [form] = Form.useForm();
  const [dataSource, setDataSource] = useState([])
  const [open, setOpen] = useState<boolean>(false);
  const [searchParams, setSearchParams] = useState({
    pageSize: 10,
    page: 1,
    query: undefined,
    total: undefined,
  })

  const [generateOpen, setGenerateOpen] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [currentRecord, setCurrentRecord] = useState()

  async function createApply(values) {
    const response: any = await StackService.applyStackAsync({
      body: {
        ...values,
      },
      query: {
        workspace: values?.workspace,
        stackID: stackId,
        type: 'Apply',
      },
      path: {
        stackID: stackId,
      }
    })
    return response
  }

  async function createGenerate(values) {
    const response: any = await StackService.generateStackAsync({
      body: {
        ...values,
      },
      query: {
        workspace: values?.workspace,
      },
      path: {
        stackID: stackId,
      }
    })
    return response
  }

  async function createDestroy(values) {
    const response: any = await StackService.destroyStackAsync({
      body: {
        ...values,
      },
      query: {
        workspace: values?.workspace,
      },
      path: {
        stackID: stackId,
      }
    })
    return response
  }

  async function createPreview(values) {
    const response: any = await StackService.previewStackAsync({
      body: {
        ...values,
      },
      query: {
        workspace: values?.workspace,
      },
      path: {
        stackID: stackId,
      }
    })
    return response
  }

  function handleSubmit(values) {
    const type = values?.type;
    let response = undefined;
    if (type === 'Apply') {
      response = createApply(values)
    } else if (type === 'Generate') {
      response = createGenerate(values)
    } else if (type === 'Destroy') {
      response = createDestroy(values)
    } else {
      response = createPreview(values)
    }
    if (response?.data?.success) {
      message.success('Create Successful')
      setOpen(false)
    } else {
      message.error(response?.data?.message)
    }
  }
  function handleClose() {
    setOpen(false)
  }

  function handleReset() {
    form.resetFields();
    setSearchParams({
      ...searchParams,
      query: undefined,
    })
    getListRun({
      page: 1,
      pageSize: 10,
      query: undefined,
    })
  }

  function handleSearch() {
    const values = form.getFieldsValue()
    setSearchParams({
      ...searchParams,
      query: values,
    })
    getListRun({
      page: 1,
      pageSize: 10,
      query: values
    })
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  async function getListRun(params) {
    try {
      let startTime, endTime
      if (params?.query?.createTime) {
        const [startDate, endDate] = params?.query?.createTime;
        startTime = dayjs(startDate).utc().format(timeFormatter)
        endTime = dayjs(endDate).utc().format(timeFormatter)
      }
      const response: any = await StackService.listRun({
        query: {
          type: params?.query?.type,
          status: params?.query?.status,
          startTime,
          endTime,
          pageSize: params?.pageSize || 10,
          page: params?.page,
          stackID: stackId
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.runs);
        setSearchParams({
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        })
      } else {
        message.error(response?.data?.messaage)
      }
    } catch (error) {
    }
  }

  useEffect(() => {
    if (stackId) {
      getListRun(searchParams)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stackId])

  function handleChangePage(page: number, pageSize: number) {
    getListRun({
      ...searchParams,
      page,
      pageSize,
    })
  }

  function handleCheckDetail(record) {
    setCurrentRecord(record)
    if (record?.type === 'Generate' || record?.type === 'Apply' || record?.type === 'Destroy') {
      setGenerateOpen(true)
    } else {
      setPreviewOpen(true)
    }

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
      render: (text) => {
        return <Tag color={text === 'Succeeded' ? 'success' : text === 'Failed' ? 'error' : 'warning'}>{RUNS_STATUS_MAP?.[text]}</Tag>
      }
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (_, record) => <Button type='link' onClick={() => handleCheckDetail(record)}>Detail</Button>
    },
  ]

  function handleCreateRuns() {
    setOpen(true)
  }

  function handlGenerateColse() {
    setGenerateOpen(false)
  }
  function handlePreviewClose() {
    setPreviewOpen(false)
  }

  function renderTableTitle() {
    return <div className={styles.project_runs_toolbar}>
      <div className={styles.project_runs_toolbar_left}>
        {
          (searchParams?.total !== null || searchParams?.total !== undefined)
            ? <div className={styles.project_runs_result}>Total<Button style={{ padding: 4 }} type='link'>{searchParams?.total}</Button>results</div>
            : null
        }
        <div className={styles.projects_content_toolbar_list}>
          {
            searchParams?.query && Object.entries(searchParams?.query)?.filter(([key, _value]) => _value)?.map(([key, __value]: any) => {
              if (key === 'createTime') {
                const [startDate, endDate] = __value;
                const startTime = dayjs(startDate).utc().format(timeFormatter)
                const endTime = dayjs(endDate).utc().format(timeFormatter)
                return (
                  <div key={key} className={styles.projects_content_toolbar_item}>
                    {key}: {`${startTime} ~ ${endTime}`}
                    <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} />
                  </div>
                )
              }
              return (
                <div key={key} className={styles.projects_content_toolbar_item}>
                  {key}: {__value}
                  <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} />
                </div>
              )
            })
          }
        </div>
        {
          Object.entries(searchParams?.query || {})?.filter(([key, val]) => val)?.length > 0 && <div className={styles.projects_content_toolbar_clear}>
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
              <Select placeholder="Please select type" style={{ width: 150 }}>
                {
                  Object.entries(RUNS_TYPES)?.map(([key, value]) => <Select.Option key={key} value={value}>{value}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item name="status" label="Status">
              <Select placeholder="Please select status" style={{ width: 150 }}>
                {
                  Object.entries(RUNS_STATUS_MAP)?.map(([key, value]) => <Select.Option key={key} value={value}>{value}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item name="createTime" label="Create Time">
              <DatePicker.RangePicker showTime={{ format: 'HH:mm' }} />
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
          rowKey="id"
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
        <RunsForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} runsTypes={RUNS_TYPES} />
        <GenerateDetail currentRecord={currentRecord} open={generateOpen} handleClose={handlGenerateColse} />
        {
          previewOpen && <PreviewDetail currentRecord={currentRecord} open={previewOpen} handleClose={handlePreviewClose} />
        }
      </div>
    </div>
  )
}

export default Runs
