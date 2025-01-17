import React, { useEffect, useRef, useState } from 'react'
import queryString from 'query-string'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button, Form, Input, message, Popconfirm, Space, Table, Select } from 'antd'
import type { TableColumnsType } from 'antd';
import { PlusOutlined } from '@ant-design/icons'
import { ModuleService } from '@kusionstack/kusion-api-client-sdk'
import DescriptionWithTooltip from '@/components/descriptionWithTooltip'
import ModuleForm from './component/moduleForm'

import styles from './styles.module.less'



const ModulePage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm();
  const [open, setOpen] = useState(false)
  const [actionType, setActionType] = useState('ADD')
  const [formData, setFormData] = useState()
  const { pageSize = 10, page = 1, total = 0, moduleName } = queryString.parse(location?.search);
  const [searchParams, setSearchParams] = useState({
    pageSize,
    page,
    query: {
      moduleName
    },
    total,
  });
  const [dataSource, setDataSource] = useState([])
  const searchParamsRef = useRef<any>();

  useEffect(() => {
    const newParams = queryString.stringify({
      moduleName,
      ...(searchParamsRef.current?.query || {}),
      page: searchParamsRef.current?.page,
      pageSize: searchParamsRef.current?.pageSize,
    })
    navigate(`?${newParams}`)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, navigate])

  async function getModuleList(params) {
    try {
      const response: any = await ModuleService.listModule({
        query: {
          moduleName: params?.query?.moduleName,
          page: params?.page || searchParams?.page,
          pageSize: params?.pageSize || searchParams?.pageSize,
        }
      });
      if (response?.data?.success) {
        setDataSource(response?.data?.data?.modules);
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
    getModuleList(searchParams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleReset() {
    form.resetFields();
    const newParams = {
      ...searchParams,
      query: {
        moduleName: undefined
      }
    }
    setSearchParams(newParams)
    searchParamsRef.current = newParams;
    getModuleList({
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
    getModuleList({
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

  async function confirmDelete(record) {
    const response: any = await ModuleService.deleteModule({
      path: {
        moduleName: record?.name
      },
    })
    if (response?.data?.success) {
      message.success('delete successful')
      getModuleList(searchParams)
    } else {
      message.error(response?.data?.message)
    }
  }



  const columns: TableColumnsType<any> = [
    {
      title: 'Name',
      dataIndex: 'name',
      width: 300,
      fixed: 'left',
    },
    {
      title: 'Registry',
      dataIndex: 'registry',

      render: (_, record) => {
        return `${record?.url?.Scheme}://${record?.url?.Host}${record?.url?.Path}`;
      }
    },
    {
      title: 'Description',
      dataIndex: 'description',

      render: (desc) => {
        return <DescriptionWithTooltip desc={desc} width={400} />
      }
    },
    {
      title: 'Action',
      dataIndex: 'action',
      fixed: 'right',
      width: 200,
      render: (_, record) => {
        return (
          <Space>
            {record?.doc?.Host ? (
              <Button style={{ padding: '0px' }} type='link' href={`${record?.doc?.Scheme}://${record?.doc?.Host}${record?.doc?.Path}`} target='_blank'>doc</Button>
            ) : (
              <Button style={{ padding: '0px' }} type='link' disabled>doc</Button>
            )}
            <span style={{ padding: '0px 10px' }}></span>
            <Button style={{ padding: '0px' }} type='link' onClick={() => handleEdit(record)}>edit</Button>
            <span>/</span>
            <Popconfirm
              title="Delete the module"
              description="Are you sure to delete this module?"
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


  useEffect(() => {
    form.setFieldsValue({
      moduleName: searchParams?.query?.moduleName
    })
  }, [searchParams?.query, form])


  async function handleSubmit(values, callback) {
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
      callback && callback()
    } else {
      message.error(response?.data?.messaage)
    }
  }

  function handleCancel() {
    setFormData(undefined)
    setOpen(false)
  }

  function handleChangePage(page: number, pageSize: any) {
    getModuleList({
      ...searchParams,
      page,
      pageSize,
    })
  }

  const sourceFormProps = {
    open,
    actionType,
    handleSubmit,
    formData,
    handleCancel,
  }

  return (
    <div className={styles.modules}>
      <div className={styles.modules_action}>
        <h3>Modules</h3>
        <div className={styles.modules_action_create}>
          <Button type="primary" onClick={handleAdd}>
            <PlusOutlined /> New Module
          </Button>
        </div>
      </div>
      <div className={styles.modules_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="moduleName" label="Module Name">
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
          title={() => <h4>Module List</h4>}
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
                    width: 60,
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
      <ModuleForm {...sourceFormProps} />
    </div>
  )
}

export default ModulePage
