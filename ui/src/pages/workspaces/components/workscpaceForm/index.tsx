import React, { useEffect, useState } from 'react';
import { Modal, Button, Form, Input, message, Select } from 'antd';
import { BackendService } from "@kusionstack/kusion-api-client-sdk"

import styles from './styles.module.less';

const WorkspaceFrom = ({ open, handleClose, handleSubmit }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const [backendList, setBackendlist] = useState([])

  const formInitialValues = {
    name: '',
    description: '',
  };

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

  // 提交表单
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

  function handleCancel() {
    form.resetFields();
    handleClose();
  }

  console.log(backendList, "===backendList===")

  return (
    <Modal
      open={open}
      title="Create New Workspace"
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
        initialValues={formInitialValues}
        layout="vertical"
      >
        <Form.Item name="backendID" label="BackendID">
          <Select placeholder="Select a backend">
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
            className={styles.inputConfigPath}
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
