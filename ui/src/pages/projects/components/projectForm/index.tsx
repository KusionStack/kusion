import React, { useState } from 'react';
import { Modal, Button, Form, Select, Input, message } from 'antd';

import styles from './styles.module.less';

const ProjectForm = ({ open, handleClose, handleSubmit }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const formInitialValues = {
    name: '',
    description: '',
    projectSource: '',
    configPath: '',
    organization: '',
  };

  const onFinish = async () => {
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
          cancel
        </Button>,
        <Button key="create" type="primary" onClick={onFinish}>
          Submit
        </Button>,
      ]}
    >
      <Form
        form={form}
        initialValues={formInitialValues}
        layout="vertical"
      >
        <Form.Item name="name" label="Name">
          <Input
            placeholder="Please input Name"
            className={styles.inputConfigPath}
          />
        </Form.Item>
        <Form.Item name="projectSource" label="Project Source">
          <Select
            placeholder="Please projects source"
            className={styles.selectInput}
          />
        </Form.Item>
        <Form.Item name="configPath" label="Config Path">
          <Input
            placeholder="Please input config path in source"
            className={styles.inputConfigPath}
          />
        </Form.Item>
        <Form.Item name="organization" label="Organization">
          <Select placeholder="Please select organization" className={styles.selectInput} />
        </Form.Item>
        <Form.Item name="description" label="Description">
          <Input.TextArea
            placeholder="Please input description..."
            rows={4}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ProjectForm;
