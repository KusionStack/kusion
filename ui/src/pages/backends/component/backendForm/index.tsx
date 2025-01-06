import React, { useEffect, useState } from 'react'
import isUrl from 'is-url'
import { Button, Modal, Form, Input, Space, message, Select } from 'antd'

const BackendForm = ({
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
      form.setFieldsValue({
        name: formData?.name,
        type: formData?.backendConfig?.type,
        description: formData?.description,
        configs: formData?.backendConfig?.configs ? JSON.stringify(formData?.backendConfig?.configs) : '',
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
      ? 'New Backend'
      : actionType === 'EDIT'
        ? 'Edit Backend'
        : 'Backend Detail'
  }

  return (
    <div>
      <Modal
        title={getTitle()}
        open={open}
        onCancel={onClose}
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
          <Form form={form} layout="vertical">
            <Form.Item label="Name" name='name' rules={[
              {
                required: true,
              },
            ]}>
              <Input />
            </Form.Item>
            <Form.Item label="Type" name='type' rules={[
              {
                required: true,
              },
            ]}>
              <Select placeholder="please select type">
                {
                  ['oss', 's3', 'local']?.map(item => <Select.Option key={item} value={item}>{item}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.Item
              label="Configs"
              name="configs"
              rules={[
                {
                  required: false,
                },
              ]}
            >
              <Input.TextArea rows={4} />
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

export default BackendForm
