import React, { useEffect } from 'react'
// import styles from './styles.module.less'
import { Button, Drawer, Form, Input, Space } from 'antd'

const SourceForm = ({
  open,
  actionType,
  handleSubmit,
  formData,
  handleCancel,
}) => {
  const [form] = Form.useForm()

  useEffect(() => {
    if (formData) {
      form.setFieldsValue(formData)
    }
  }, [formData, form])

  function onSubmit() {
    handleSubmit()
  }

  function onClose() {
    form.resetFields()
    handleCancel()
  }

  function getTitle() {
    return actionType === 'ADD'
      ? '新增Source'
      : actionType === 'EDIT'
        ? '编辑Source'
        : 'Source详情'
  }

  return (
    <div>
      <Drawer
        title={getTitle()}
        open={open}
        onClose={onClose}
        extra={
          <Space>
            <Button onClick={onClose}>Cancel</Button>
            <Button onClick={onSubmit} type="primary">
              Submit
            </Button>
          </Space>
        }
      >
        <Form form={form} layout="vertical">
          <Form.Item label="Name" name={'name'}>
            <Input />
          </Form.Item>
          <Form.Item label="Url" name={'url'}>
            <Input />
          </Form.Item>
        </Form>
      </Drawer>
    </div>
  )
}

export default SourceForm
