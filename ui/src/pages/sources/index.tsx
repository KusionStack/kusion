import React, { useEffect, useRef, useState } from 'react'
import queryString from 'query-string'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button, Form, Input, message, Popconfirm, Space, Table, Select } from 'antd'
import type { TableColumnsType } from 'antd';
import { PlusOutlined } from '@ant-design/icons'
import { SourceService } from '@kusionstack/kusion-api-client-sdk'
import SourceForm from './component/sourceForm'
import DescriptionWithTooltip from '@/components/descriptionWithTooltip'

import styles from './styles.module.less'



const SourcesPage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const { pageSize = 10, page = 1, total = 0, sourceName } = queryString.parse(location?.search);
  const [searchParams, setSearchParams] = useState({
    pageSize,
    page,
    query: {
      sourceName
    },
    total,
  });

  const [dataSource, setDataSource] = useState([])
  const searchParamsRef = useRef<any>();

  useEffect(() => {
    const newParams = queryString.stringify({
      sourceName,
      ...(searchParamsRef.current?.query || {}),
      page: searchParamsRef.current?.page,
      pageSize: searchParamsRef.current?.pageSize,
    })
    navigate(`?${newParams}`)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, navigate])

  async function getResourceList(params) {
    try {
      const response: any = await SourceService.listSource({
        query: {
          sourceName: params?.query?.sourceName,
          page: params?.page || searchParams?.page,
          pageSize: params?.pageSize || searchParams?.pageSize,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.sources);
        const newParams = {
          query: params?.query,
          pageSize: response?.data?.data?.pageSize,
          page: response?.data?.data?.currentPage,
          total: response?.data?.data?.total,
        }
        setSearchParams(newParams)
        searchParamsRef.current = newParams;
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    getResourceList(searchParams)
  }, [])

  function handleReset() {
    form.resetFields();
    const newParams = {
      ...searchParams,
      query: {
        sourceName: undefined
      }
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getResourceList({
      page: 1,
      pageSize: 10,
      query: undefined
    })
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    const newParams = {
      ...searchParams,
      query: values
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getResourceList({
      page: 1,
      pageSize: 10,
      query: values,
    })
  }

  function handleAdd() {
    setActionType('ADD')
    setOpen(true)
  }
  function handleEdit(record) {
    setActionType('EDIT')
    setOpen(true)
    setFormData(record)
  }

  function handleChangePage(page: number, pageSize: any) {
    getResourceList({
      ...searchParams,
      page,
      pageSize,
    })
  }

  useEffect(() => {
    form.setFieldsValue({
      sourceName: searchParams?.query?.sourceName
    })
  }, [searchParams?.query, form])

  async function confirmDelete(record) {
    const response: any = await SourceService.deleteSource({
      path: {
        sourceID: record?.id
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getResourceList(searchParams)
    } else {
      message.error(response?.data?.message)
    }
  }



  const columns: TableColumnsType<any> = [
    {
      title: 'Name',
      dataIndex: 'name',
      width: 300,
      fixed: 'left'
    },
    {
      title: 'Description',
      dataIndex: 'description',
      render: (desc) => {
        return <DescriptionWithTooltip desc={desc} width={450} />
      }
    },
    {
      title: 'Url',
      dataIndex: 'remote',
      render: (remoteObj) => {
        return `${remoteObj?.Scheme}://${remoteObj?.Host}${remoteObj?.Path}`
      }
    },
    {
      title: 'Action',
      dataIndex: 'action',
      fixed: 'right',
      width: 150,
      render: (_, record) => {
        return (
          <Space>
            <Button style={{ padding: '0px' }} type='link' onClick={() => handleEdit(record)}>edit</Button>
            <span>/</span>
            <Popconfirm
              title="Delete the source"
              description="Are you sure to delete this source?"
              onConfirm={() => confirmDelete(record)}
              okText="Yes"
              cancelText="No"
            >
              <Button style={{ padding: '0px' }} type='link' danger>delete</Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  async function handleSubmit(values, callback) {
    let response: any
    if (actionType === 'EDIT') {
      response = await SourceService.updateSource({
        body: {
          id: (formData as any)?.id,
          name: values?.name,
          remote: values?.remote,
          sourceProvider: values?.sourceProvider,
          description: values?.description
        },
        path: {
          sourceID: (formData as any)?.id
        }
      })
    } else {
      response = await SourceService.createSource({
        body: {
          name: values?.name,
          remote: values?.remote,
          sourceProvider: values?.sourceProvider,
          description: values?.description,
        }
      })
    }

    if (response?.data?.success) {
      message.success(actionType === 'EDIT' ? 'Update Successful' : 'Create Successful')
      getResourceList(searchParams)
      callback && callback()
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
    <div className={styles.sources}>
      <div className={styles.sources_action}>
        <h3>Sources</h3>
        <div className={styles.sources_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Source
          </Button>
        </div>
      </div>
      <div className={styles.sources_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="sourceName" label="Source Name">
              <Input />
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
      <div className={styles.modules_content}>
        <Table
          title={() => <h4>Source List</h4>}
          rowKey="id"
          columns={columns}
          scroll={{ x: 1300 }}
          dataSource={dataSource}
          pagination={{
            total: Number(searchParams?.total),
            current: Number(searchParams?.page),
            pageSize: Number(searchParams?.pageSize),
            showTotal: (total, range) => (
              <div style={{
                fontSize: '12px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'flex-end'
              }}>
                show{' '}
                <Select
                  value={searchParams?.pageSize}
                  size="small"
                  style={{
                    width: 55,
                    margin: '0 4px',
                    fontSize: '12px'
                  }}
                  onChange={(value) => handleChangePage(1, value)}
                  options={['10', '15', '20', '30', '40', '50', '75', '100'].map((value) => ({ value, label: value }))}
                />
                items, {range[0]}-{range[1]} of {total} items
              </div>
            ),
            size: "default",
            style: {
              marginTop: '16px',
              textAlign: 'right'
            },
            onChange: (page, size) => {
              handleChangePage(page, size);
            },
          }}
        />
      </div>
      <SourceForm {...sourceFormProps} />
    </div>
  )
}

export default SourcesPage
