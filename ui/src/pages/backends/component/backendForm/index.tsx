import React, { useEffect, useState } from 'react'
import { Button, Modal, Form, Input, Space, Select } from 'antd'
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

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
      const configs = formData?.backendConfig?.configs ? Object.entries(formData?.backendConfig?.configs)?.map(([key, value]) => ({ key, value })) : []
      form.setFieldsValue({
        name: formData?.name,
        type: formData?.backendConfig?.type,
        description: formData?.description,
        configs: configs
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
      const configObj = {};
      values?.configs?.forEach(({ key, value }) => {
        if (key && value) {
          configObj[key] = value
        }
      })
      handleSubmit({
        ...values,
        configs: Object.keys(configObj)?.length > 0 ? configObj : undefined
      }, () => {
        form.resetFields()
      })
    } catch (e) {
      console.log(e, "Error")
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
        width={650}
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
          <Form form={form} layout="vertical"
            initialValues={{
              configs: [{
                key: undefined,
                value: undefined
              }]
            }}
          >
            <Form.Item label="Name" name='name' rules={[
              {
                required: true,
              },
            ]}>
              <Input placeholder="Enter backend name" />
            </Form.Item>
            <Form.Item label="Type" name='type' rules={[
              {
                required: true,
              },
            ]}>
              <Select placeholder="Select backend type">
                {
                  [
                    { label: 'aliyun-oss', value: 'oss' },
                    { label: 'aws-s3', value: 's3' },
                    { label: 'local', value: 'local' },
                    { label: 'google-cloud-storage', value: 'google' }
                  ].map(item => <Select.Option key={item.value} value={item.value}>{item.label}</Select.Option>)
                }
              </Select>
            </Form.Item>
            <Form.List name="configs">
              {(fields, { add, remove }) => {
                return (
                  <>
                    {fields.map(({ key, name, ...restField }, index) => {
                      return (
                        <Form.Item
                          label={index === 0 ? 'Configs' : ''}
                          key={key}
                          style={{ marginBottom: 0 }}
                        >
                          <Space
                            style={{ display: 'flex', marginBottom: 8, width: '100%' }}
                            align="baseline"
                          >
                            <Form.Item
                              style={{ flex: 1 }}
                              {...restField}
                              name={[name, 'key']}
                              rules={[
                                {
                                  required: false,
                                  message: 'Missing first name',
                                },
                              ]}
                            >
                              <Input style={{ width: 250 }} placeholder="Enter key" />
                            </Form.Item>
                            <Form.Item
                              style={{ flex: 1 }}
                              {...restField}
                              name={[name, 'value']}
                              rules={[{ required: false, message: 'Required' }]}
                            >
                              <Input style={{ width: 250 }} placeholder="Enter value" />
                            </Form.Item>
                            {fields?.length > 1 && (
                              <MinusCircleOutlined onClick={() => remove(name)} />
                            )}
                          </Space>
                        </Form.Item>
                      )
                    })}
                    <Form.Item>
                      <Button
                        type="dashed"
                        onClick={() => add()}
                        block
                        icon={<PlusOutlined />}
                      >
                        Add
                      </Button>
                    </Form.Item>
                  </>
                )
              }}
            </Form.List>
            <Form.Item name="description" label="Description"
              getValueFromEvent={(e) => {
                const currentValue = e.target.value;
                const previousValue = form.getFieldValue('description') || '';
                const wordCount = currentValue.trim().split(/\s+/).filter(Boolean).length;

                // If word count exceeds 200, return the previous value
                return wordCount <= 200 ? currentValue : previousValue;
              }}
            >
              <Input.TextArea
                placeholder="Enter description for this backend..."
                rows={4}
                showCount={{
                  formatter: ({ value }) => {
                    const words = value ? value.trim().split(/\s+/).filter(Boolean).length : 0;
                    return `${words} / 200 words`;
                  }
                }}
              />
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </div>
  )
}

export default BackendForm
