import React, { useEffect, useState } from 'react'
import { Button, Modal, Form, Input, Space, message } from 'antd'
import isUrl from 'is-url'

const ModuleForm = ({
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
      const url = formData?.url;
      const dovUrl = formData?.doc;
      form.setFieldsValue({
        ...formData,
        url: `${url?.Scheme}${url?.Host}${url?.Path}`,
        doc: `${dovUrl?.Scheme}//${dovUrl?.Host}${dovUrl?.Path}`,
      })
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
      ? 'New Module'
      : 'Edit Module'
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
            <Input disabled={actionType === 'EDIT'} />
          </Form.Item>
          <Form.Item
            label="url"
            name="url"
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
            name="doc"
            rules={[
              {
                required: false,
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Description"
            name="description"
            rules={[
              {
                required: false,
              },
            ]}
          >
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ModuleForm
