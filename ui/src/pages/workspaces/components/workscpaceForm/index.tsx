import React, { useEffect, useState } from 'react';
import { Modal, Button, Form, Input, message, Select } from 'antd';
import { BackendService } from "@kusionstack/kusion-api-client-sdk"

const WorkspaceFrom = ({ open, actionType, handleClose, handleSubmit, formData }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const [backendList, setBackendlist] = useState([])

  useEffect(() => {
    if (formData) {
      form.setFieldsValue({
        name: formData?.name,
        description: formData?.description,
        backendID: formData?.backend?.name
      })
    }
  }, [formData, form])

  async function getBackendList() {
    const response: any = await BackendService.listBackend()
    if (response?.data?.success) {
      setBackendlist(response?.data?.data?.backends)
    } else {
      message.error(response?.data?.message || 'request error')
    }
  }

  useEffect(() => {
    getBackendList()
  }, [])

  // Submit form
  const onFinish = async () => {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      const values = form.getFieldsValue();
      handleSubmit(values,() => {
        form.resetFields()
      })
    } catch (e) {
      message.error('submit failed');
    } finally {
      setLoading(false);
    }
  };

  function handleCancel() {
    form.resetFields();
    handleClose();
  }
  
  function getTitle() {
    return actionType === 'ADD'
      ? 'New Workspace'
      : actionType === 'EDIT'
        ? 'Edit Workspace'
        : 'Workspace Detail'
  }

  return (
    <Modal
      open={open}
      title={getTitle()}
      footer={[
        <Button key="cancel" onClick={handleCancel}>
          Cancel
        </Button>,
        <Button key="create" type="primary" onClick={onFinish}>
          Submit
        </Button>,
      ]}
      onCancel={handleCancel}
    >
      <Form
        form={form}
        layout="vertical"
      >
        <Form.Item name="backendID" label="Backend">
          <Select placeholder="Select a backend" disabled={actionType === 'EDIT'}>
            {
              backendList?.map((item: any) => {
                return <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>
              })
            }
          </Select>
        </Form.Item>
        <Form.Item name="name" label="Name">
          <Input
            placeholder="Enter workspace name"
          />
        </Form.Item>
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
            placeholder="Enter description for this workspace..."
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
    </Modal>
  );
};

export default WorkspaceFrom;
