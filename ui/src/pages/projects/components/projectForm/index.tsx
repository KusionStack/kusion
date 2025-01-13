import React, { useState } from 'react';
import { Modal, Button, Form, Select, Input, message } from 'antd';

const ProjectForm = ({ open, handleClose, handleSubmit, sourceList }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const onFinish = async () => {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      const values = form.getFieldsValue();
      handleSubmit(values)
    } catch (e) {
      message.error('submit failed');
    } finally {
      setLoading(false);
    }
  };

  function onClose() {
    form.resetFields()
    handleClose()
  }

  return (
    <Modal
      open={open}
      title="Create Project"
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button key="create" type="primary" onClick={onFinish}>
          Submit
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout="vertical"
      >
        <Form.Item name="name" label="Name"
          rules={[
            {
              required: true,
            },
            {
              validator: (_, value) => {
                if (!value) {
                  return Promise.reject('Name is required');
                }
                const nameRegex = /^[a-zA-Z0-9_-]+$/;
                if (nameRegex.test(value)) {
                  return Promise.resolve();
                }
                return Promise.reject('Name can only contain letters, numbers, underscores and hyphens');
              }
            }
          ]}
        >
          <Input
            placeholder="Enter project name"
          />
        </Form.Item>
        <Form.Item name="projectSource" label="Project Source"
          rules={[
            {
              required: true,
            },
          ]}
        >
          <Select
            placeholder="Select project source"
          >
            {
              sourceList?.map(item => {
                return <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>
              })
            }
          </Select>
        </Form.Item>
        <Form.Item name="path" label="Path"
          rules={[
            {
              required: true,
            },
            {
              validator: (_, value) => {
                if (!value) {
                  return Promise.reject('Required')
                }
                const pathRegex = /^[a-zA-Z0-9_\/-]+$/;
                if (value.startsWith('/')) {
                  return Promise.reject('Path should be relative (without leading slash)')
                }
                if (pathRegex.test(value)) {
                  return Promise.resolve()
                }
                return Promise.reject('Invalid path format')
              },
            },
          ]}
        >
          <Input
            placeholder="Enter path from source root (e.g. path/to/project)"
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
            placeholder="Enter description for this project..."
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

export default ProjectForm;
