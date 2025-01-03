import React, { useEffect, useState } from 'react'
import isUrl from 'is-url'
import { Button, Modal, Form, Input, Space, message, Select } from 'antd'

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
      const remote = formData?.remote;
      form.setFieldsValue({
        name: formData?.name,
        sourceProvider: formData?.sourceProvider,
        description: formData?.description,
        remote: `${remote?.Scheme}//${remote?.Host}${remote?.Path}`,
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
            <Form.Item label="Name" name='name' rules={[
              {
                required: true,
              },
            ]}>
              <Input />
            </Form.Item>
            <Form.Item label="Remote" name='remote' rules={[
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
            ]}>
              <Input />
            </Form.Item>
            <Form.Item label="SourceProvider" name='sourceProvider'>
              <Select placeholder="Please select source provider">
                <Select.Option key="git" value="git">git</Select.Option>
              </Select>
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
        </div>
      </Modal>
    </div>
  )
}

export default SourceForm
