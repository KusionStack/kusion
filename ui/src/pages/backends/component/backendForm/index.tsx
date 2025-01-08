import React, { useEffect, useState } from 'react'
import { Button, Modal, Form, Input, Space, message, Select } from 'antd'
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
              <Input />
            </Form.Item>
            <Form.Item label="Type" name='type' rules={[
              {
                required: true,
              },
            ]}>
              <Select placeholder="please select type">
                {
                  ['oss', 's3', 'local', 'google']?.map(item => <Select.Option key={item} value={item}>{item}</Select.Option>)
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
                          required={true}
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
                              <Input style={{ width: 250 }} />
                            </Form.Item>
                            <Form.Item
                              style={{ flex: 1 }}
                              {...restField}
                              name={[name, 'value']}
                              rules={[{ required: false, message: 'Required' }]}
                            >
                              <Input style={{ width: 250 }} />
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
