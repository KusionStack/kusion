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
          <Select placeholder="Please select a backend">
            {
              backendList?.map((item: any) => {
                return <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>
              })
            }
          </Select>
        </Form.Item>
        <Form.Item name="name" label="Name">
          <Input
            placeholder="Please input"
            className={styles.inputConfigPath}
          />
        </Form.Item>
        <Form.Item name="description" label="Description">
          <Input.TextArea
            placeholder="This is a description, it may be long or short..."
            rows={4}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default WorkspaceFrom;
