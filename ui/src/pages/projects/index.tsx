import React, { useEffect, useState } from 'react'
import styles from "./styles.module.less"
import { Button, Card, Col, DatePicker, Form, Input, Row, Space, Table, Tag } from 'antd'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { ProjectService } from '@kusionstack/kusion-api-client-sdk'
import ProjectForm from './components/projectForm'

type TypeSearchParams = {
  name?: string
  org?: string
  creator?: string
  createTime?: string
} | undefined

const Projects = () => {
  const [form] = Form.useForm();
  const [searchParams, setSearchParams] = useState<TypeSearchParams>();
  const [dataSource, setDataSource] = useState([])
  const [open, setOpen] = useState<boolean>(false);

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
    setSearchParams(undefined)
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    setSearchParams(values)
    console.log(values, "=====handleSearch=====")
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    // setSearchParams({
    //   ...searchParams,
    //   [key]: ''
    // })
    handleSearch()
  }

  async function getList(params) {
    try {
      const resData: any = await ProjectService.listProject(params);
      setDataSource(resData?.data?.data);
      console.log(resData, "======resData====")
    } catch (error) {

    }
  }

  useEffect(() => {
    getList({})
  }, [])


  const colums = [
    {
      title: 'Name',
      dataIndex: 'name',
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: 'Org',
      dataIndex: 'org',
    },
    {
      title: 'Creator',
      dataIndex: 'creator',
    },
    {
      title: 'Create Time',
      dataIndex: 'creationTimestamp',
    },
  ]

  function renderTableTitle(currentPageData) {
    console.log(currentPageData, "=====currentPageData====")
    return <div className={styles.projects_content_toolbar}>
      <h4>Project List</h4>
      <div className={styles.projects_content_toolbar_list}>
        {
          searchParams && Object.entries(searchParams)?.filter(([key, value]) => value)?.map(([key, value]) => {
            return <div className={styles.projects_content_toolbar_item}>{key}: {value} <CloseOutlined style={{ marginLeft: 10, color: '#140e3540' }} onClick={() => handleClear(key)} /></div>
          })
        }
      </div>
      {
        searchParams && <div className={styles.projects_content_toolbar_clear}>
          <Button type='link' onClick={handleReset} style={{ paddingLeft: 0 }}>Clear</Button>
        </div>
      }
    </div>
  }

  return (
    <div className={styles.projects}>
      <div className={styles.projects_action}>
        <h3>Projects</h3>
        <div className={styles.projects_action_create}>
          <Button type='primary' onClick={handleCreate}><PlusOutlined /> Create Projects</Button>
        </div>
      </div>
      {/* Search Form block*/}
      <div className={styles.projects_search}>
        <Form form={form} style={{ marginBottom: 0 }}>
          <Space>
            <Form.Item name="name" label="Project Name">
              <Input />
            </Form.Item>
            <Form.Item name="org" label="Org">
              <Input />
            </Form.Item>
            <Form.Item name="creator" label="Creator">
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
      <div className={styles.projects_content}>
        <Table title={renderTableTitle} columns={colums} dataSource={dataSource} />
      </div>
      <ProjectForm open={open} handleSubmit={handleSubmit} handleClose={handleClose} />
    </div>
  )
}

export default Projects
