import React, { useEffect, useState } from 'react'
// import styles from './styles.module.less'
import { Button, Modal, Form, Input, Space, message } from 'antd'

const SourceForm = ({
  open,
  actionType,
  handleSubmit,
  formData,
  handleCancel,
}) => {
  const [form] = Form.useForm()

  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (formData) {
      form.setFieldsValue(formData)
    }
  }, [formData, form])

  function onSubmit() {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      const values = form.getFieldsValue();
      handleSubmit(values)
    } catch (e) {
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
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
            <Form.Item label="Name" name='name'>
              <Input />
            </Form.Item>
            <Form.Item label="Remote" name='remote'>
              <Input />
            </Form.Item>
            <Form.Item label="SourceProvider" name='sourceProvider'>
              <Input />
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </div>
  )
}

export default SourceForm
