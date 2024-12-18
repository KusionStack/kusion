import React, { useEffect } from 'react'
// import styles from './styles.module.less'
import { Button, Modal, Form, Input, Space } from 'antd'
import isUrl from 'is-url'

const ModuleForm = ({
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
        width={560}
        title={getTitle()}
        open={open}
        onClose={onClose}
        onCancel={onClose}
        footer={
          [
            <Space>
              <Button onClick={onClose}>Cancel</Button>
              <Button onClick={onSubmit} type="primary">
                Submit
              </Button>
            </Space>
          ]
        }
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="Name"
            name="name"
            rules={[
              {
                required: true,
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Registry"
            name="registry"
            rules={[
              {
                required: true,
              },
              {
                validator: (_, value) => {
                  if (!value) {
                    return Promise.reject('必填项')
                  }
                  if (isUrl(value)) {
                    return Promise.resolve()
                  } else {
                    return Promise.reject('不是一个URL')
                  }
                },
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Document Url"
            name="documentUrl"
            rules={[
              {
                required: true,
              },
            ]}
          >
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ModuleForm
