import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { Button, DatePicker, Form, message, Space, Table, Tag, Select, Tooltip } from 'antd'
import type { TableColumnsType } from 'antd'
import { CloseOutlined, PlusOutlined, RedoOutlined } from '@ant-design/icons'
import { useLocation, useNavigate } from 'react-router-dom'
import queryString from 'query-string'
import moment from "moment"
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import { RUNS_STATUS_MAP, RUNS_TYPES } from '@/utils/constants'
import GenerateDetail from './generateDetail'
import PreviewDetail from './previewDetail'
import RunsForm from './runsForm'
import { createApply, createDestroy, createGenerate, createPreview, queryListRun } from './service'

import styles from "./styles.module.less"

// use plugin
dayjs.extend(utc);
dayjs.extend(timezone);

const timeFormatter = 'YYYY-MM-DDTHH:mm:ssZ'

type IProps = {
  stackId: any
  panelActiveKey?: any
}

const Runs = forwardRef(({ panelActiveKey }: IProps, funcRef) => {
  const [form] = Form.useForm();
  const navigate = useNavigate()
  const location = useLocation();
  const { projectName, stackId } = queryString.parse(location?.search);
  const [dataSource, setDataSource] = useState(null)
  const [open, setOpen] = useState<boolean>(false);
  const [searchParams, setSearchParams] = useState<any>({})
  const [generateOpen, setGenerateOpen] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [currentRecord, setCurrentRecord] = useState()
  const searchParamsRef = useRef<any>();

  useEffect(() => {
    const params: any = new URLSearchParams(location.search);
    const search: any = {
      page: 1,
      pageSize: 10,
    };
    for (const [key, value] of params.entries()) {
      search[key] = value;
    }
    setSearchParams(search)
    searchParamsRef.current = search;
    const startTime = search.startTime;
    const endTime = search?.endTime;
    const timeObj = (startTime && endTime ? {
      createTime: [moment(startTime), moment(endTime as any),]
    } : {})
    form.setFieldsValue({
      type: search?.type,
      status: search?.status,
      ...timeObj,
    })
  }, [location.search, form]);



  function updateURL(paramsData) {
    const paramObj = {
      projectName,
      type: paramsData?.type,
      status: paramsData?.status,
      startTime: paramsData?.startTime,
      endTime: paramsData?.endTime,
      stackId: paramsData?.stackId || stackId,
      page: paramsData?.page,
      pageSize: paramsData?.pageSize,
      panelKey: panelActiveKey,
      sortBy: paramsData?.sortBy,
      ascending: paramsData?.ascending,
    }
    getListRun(paramObj)
    const newParams = queryString.stringify(paramObj)
    navigate({ search: newParams.toString() });
  }

  useImperativeHandle(funcRef, () => ({
    updateRunsURL: (data) => {
      updateURL({
        ...searchParamsRef.current,
        ...data,
      })
    },
  }))


  useEffect(() => {
    let timer;
    if (searchParamsRef.current) {
      timer = setInterval(() => {
        getListRun(searchParamsRef.current)
      }, 7000)
    }
    return () => {
      if (timer) clearInterval(timer)
    }
  }, [])

  useEffect(() => {
    if (stackId) {
      updateURL({
        ...searchParamsRef.current,
        stackId: Number(stackId),
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])


  async function handleSubmit(values, callback) {
    const type = values?.type;
    let response = undefined;
    if (type === 'Apply') {
      response = await createApply(values, stackId)
    } else if (type === 'Generate') {
      response = await createGenerate(values, stackId)
    } else if (type === 'Destroy') {
      response = await createDestroy(values, stackId)
    } else {
      response = await createPreview(values, stackId)
    }
    if (response?.data?.success) {
      message.success('Create Successful')
      setOpen(false)
      callback && callback()
      getListRun(searchParams)
    } else {
      message.error(response?.data?.message)
    }
  }

  function handleClose() {
    setOpen(false)
  }

  function handleSearch() {
    const values = form.getFieldsValue()
    let startTime, endTime
    if (values?.createTime) {
      const [startDate, endDate] = values?.createTime;
      startTime = dayjs(startDate).utc().format(timeFormatter)
      endTime = dayjs(endDate).utc().format(timeFormatter)
    }
    setSearchParams((prev) => {
      const updatedParams = {
        ...prev,
        type: values?.type,
        status: values?.status,
        startTime,
        endTime,
      };
      updateURL(updatedParams)
      searchParamsRef.current = updatedParams
      return updatedParams;
    })
  }

  function handleReset() {
    form.resetFields();
    handleSearch()
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  async function getListRun(params) {
    try {
      const response: any = await queryListRun({
        type: params?.type,
        status: params?.status,
        startTime: params?.startTime,
        endTime: params?.endTime,
        pageSize: params?.pageSize,
        page: params?.page,
        stackID: params?.stackId,
        sortBy: params?.sortBy,
        ascending: params?.ascending,
      })
      if (response?.data?.success) {
        setDataSource(response?.data?.data);
      } else {
        message.error(response?.data?.messaage)
      }
    } catch (error) {
    }
  }



  function handleChangePage({ current, pageSize }, filters, { field, order }) {
    setSearchParams((prev) => {
      const updatedParams = {
        ...prev,
        page: current,
        pageSize,
        sortBy: field === 'creationTimestamp' ? 'createTimestamp' : field,
        ascending: order === "ascend",
      };
      updateURL(updatedParams)
      return updatedParams;
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


  const columns: TableColumnsType<any> = [
    {
      title: 'Runs ID',
      dataIndex: 'id',
      fixed: 'left',
    },
    {
      title: 'Type',
      dataIndex: 'type',
    },
    {
      title: 'Create Time',
      dataIndex: 'creationTimestamp',
      sorter: true,
      sortDirections: ['ascend', 'descend', 'ascend'],
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
      fixed: 'right',
      width: 150,
      render: (_, record) => <Button style={{ padding: 0 }} type='link' onClick={() => handleCheckDetail(record)}>Detail</Button>
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
    const queryParams = {
      type: searchParams?.type,
      status: searchParams?.status,
      createTime: searchParams?.startTime && searchParams?.endTime ? [searchParams?.startTime, searchParams?.endTime] : undefined,
    }
    return (
      <div className={styles.project_runs_toolbar}>
        <div className={styles.project_runs_toolbar_left}>
          {
            (dataSource?.total !== null || dataSource?.total !== undefined)
              ? <div className={styles.project_runs_result}>Total<Button style={{ padding: 4 }} type='link'>{dataSource?.total}</Button>results</div>
              : null
          }
          <div className={styles.projects_content_toolbar_list}>
            {
              queryParams && Object.entries(queryParams)?.filter(([key, _value]) => _value)?.map(([key, __value]: any) => {
                if (key === 'createTime') {
                  const [startDate, endDate] = __value;
                  // const startTime = dayjs(startDate).utc().format(timeFormatter)
                  // const endTime = dayjs(endDate).utc().format(timeFormatter)
                  return (
                    <div key={key} className={styles.projects_content_toolbar_item}>
                      {key}: {`${startDate} ~ ${endDate}`}
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
            Object.entries(queryParams || {})?.filter(([key, val]) => val)?.length > 0 && (
              <div className={styles.projects_content_toolbar_clear}>
                <Button type='link' onClick={handleReset} style={{ paddingLeft: 0 }}>Clear</Button>
              </div>
            )
          }
        </div>
        <div className={styles.projects_content_toolbar_create}>
          <Space>
            <Tooltip title={'Refresh'}>
              <Button
                style={{ color: '#646566', fontSize: 18 }}
                icon={<RedoOutlined />}
                onClick={() => getListRun(searchParams)}
                type="text"
              />
            </Tooltip>
            <Button type="primary" onClick={handleCreateRuns}>
              <PlusOutlined /> New Runs
            </Button>
          </Space>
        </div>
      </div>
    )
  }


  return (
    <div className={styles.project_runs}>
      <div className={styles.project_runs_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="type" label="Type">
              <Select placeholder="Please select type" style={{ width: 150 }} allowClear>
                {
                  Object.entries(RUNS_TYPES)?.map(([key, value]) => <Select.Option key={key} value={value}>{value}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item name="status" label="Status">
              <Select placeholder="Please select status" style={{ width: 150 }} allowClear>
                {
                  Object.entries(RUNS_STATUS_MAP)?.map(([key, value]) => <Select.Option key={key} value={value}>{value}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item name="createTime" label="Create Time">
              <DatePicker.RangePicker allowClear showTime={{ format: 'HH:mm' }} />
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
      <div className={styles.project_runs_content}>
        {renderTableTitle()}
        <Table
          rowKey="id"
          columns={columns}
          dataSource={dataSource?.runs || []}
          scroll={{ x: 1300 }}
          onChange={handleChangePage}
          pagination={{
            total: dataSource?.total,
            current: Number(searchParams?.page),
            pageSize: Number(searchParams?.pageSize),
            showSizeChanger: true,
            pageSizeOptions: [10, 15, 20, 30, 40, 50, 75, 100],
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} items`,
            size: "default",
            style: {
              marginRight: 16,
              textAlign: 'right'
            },
          }}
        />
        <RunsForm
          open={open}
          handleSubmit={handleSubmit}
          handleClose={handleClose}
          runsTypes={RUNS_TYPES}
        />
        <GenerateDetail
          currentRecord={currentRecord}
          open={generateOpen}
          handleClose={handlGenerateColse}
        />
        {
          previewOpen && (
            <PreviewDetail
              currentRecord={currentRecord}
              open={previewOpen}
              handleClose={handlePreviewClose}
            />
          )
        }
      </div>
    </div>
  )
})

export default Runs


