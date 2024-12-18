import React, { useEffect } from 'react'
// import styles from './styles.module.less'
import { Button, Modal, Form, Input, Space } from 'antd'

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
      ? 'New Source'
      : actionType === 'EDIT'
        ? 'Edit Source'
        : 'Source Detail'
  }

  return (
    <div>
      <Modal
        title={getTitle()}
        open={open}
        onClose={onClose}
        footer={
          <Space>
            <Button onClick={onClose}>Cancel</Button>
            <Button onClick={onSubmit} type="primary">
              Submit
            </Button>
          </Space>
        }
      >
        <div style={{ margin: 20 }}>
          <Form form={form} layout="horizontal">
            <Form.Item label="Name" name={'name'}>
              <Input />
            </Form.Item>
            <Form.Item label="Url" name={'url'}>
              <Input />
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </div>
  )
}

export default SourceForm
