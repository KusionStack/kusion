import React, { useEffect, useState } from 'react'
import { Button, Modal, Form, Input, Space, message } from 'antd'
import isUrl from 'is-url'
import styles from './styles.module.less';

const ModuleForm = ({
  open,
  actionType,
  handleSubmit,
  formData,
  handleCancel,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false);
  
  // Listen to the form values
  const nameValue = Form.useWatch('name', form);
  const urlValue = Form.useWatch('url', form);
    
  useEffect(() => {
    if (formData) {
      const url = formData?.url;
      const docUrl = formData?.doc;
      form.setFieldsValue({
        ...formData,
        url: `${url?.Scheme}://${url?.Host}${url?.Path}`,
        doc: `${docUrl?.Scheme}://${docUrl?.Host}${docUrl?.Path}`,
      })
    }
  }, [formData, form])

  async function onSubmit() {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      await form.validateFields();
      const values = form.getFieldsValue();
      handleSubmit(values)
    } catch (e) {
      message.error('submit failed');
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
              <Button 
                onClick={onSubmit} 
                type="primary"
                disabled={!nameValue || 
                  !urlValue ||
                  form.getFieldError('name').length > 0 ||
                  form.getFieldError('url').length > 0 ||
                  form.getFieldError('doc').length > 0 ||
                  !!form.getFieldsError().filter(({errors}) => errors.length).length}
              >
                Submit
              </Button>
              {/* debug
              <div style={{marginTop: '10px', fontSize: '12px', color: '#666'}}>
                <div>Fields Touched: {JSON.stringify(form.isFieldsTouched(['name', 'url']))}</div>
                <div>Form Errors: {JSON.stringify(form.getFieldsError())}</div>
                <div>Form Values: {JSON.stringify(form.getFieldValue('name'))}</div>
              </div> */}
            </Space>
          ]
        }
      >
        <Form 
          form={form} 
          layout="vertical"
          validateTrigger={['onChange', 'onBlur']}
        >
          <Form.Item
            label="Name"
            name="name"
            rules={[
              {
                required: true,
                message: 'Please input module name',
              },
            ]}
          >
            <Input disabled={actionType === 'EDIT'} />
          </Form.Item>
          <Form.Item
            label="URL"
            name="url"
            validateTrigger={['onChange', 'onBlur']}
            rules={[
              {
                required: true,
              },
              {
                validator: (_, value) => {
                  if (!value) {
                    return Promise.reject('URL is required')
                  }
                  if (isUrl(value)) {
                    return Promise.resolve()
                  } else {
                    return Promise.reject('Please input valid URL')
                  }
                },
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Document URL" 
            name="doc"
            validateTrigger={['onChange', 'onBlur']}
            rules={[
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  if (isUrl(value)) {
                    return Promise.resolve()
                  } else {
                    return Promise.reject('Please input valid URL')
                  }
                },
              },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="Description"
            name="description"
          >
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ModuleForm
