import React, { useEffect, useState } from 'react'
// import styles from './styles.module.less'
import { Button, Modal, Form, Input, Space, message } from 'antd'

const StackForm = ({
  stackFormOpen,
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
      message.error('Submit failed');
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
      ? 'New Stack'
      : actionType === 'EDIT'
        ? 'Edit Stack'
        : 'Stack Detail'
  }

  return (
    <div>
      <Modal
        width={560}
        title={getTitle()}
        open={stackFormOpen}
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
            label="Path"
            name="path"
            rules={[
              {
                required: true,
              },
              {
                validator: (_, value) => {
                  if (!value) {
                    return Promise.reject('Required')
                  }
                  const pathRex1 = new RegExp("^(/[^/\0]+)*$");
                  const pathRex2 = new RegExp("^(\/?[^/\0]+)+$");
                  if (pathRex1.test(value) || pathRex2.test(value)) {
                    return Promise.resolve()
                  } else {
                    return Promise.reject('Not a path')
                  }
                },
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

export default StackForm
