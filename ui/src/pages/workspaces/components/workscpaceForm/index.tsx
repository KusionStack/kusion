import React, { useState } from 'react';
import { Modal, Button, Form, Input, message } from 'antd';

import styles from './styles.module.less';

const WorkscpaceForm = ({ open, handleClose, handleSubmit }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const formInitialValues = {
    name: '',
    description: '',
  };

  // 提交表单
  const onFinish = async () => {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      const values = form.getFieldsValue();
      console.log(values, "======values======")
      handleSubmit(values)
    } catch (e) {
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
  };

  function handleCancel() {
    form.resetFields();
    handleClose();
  }

  return (
    <Modal
      open={open}
      title="创建新任务"
      footer={[
        <Button key="cancel" onClick={handleCancel}>
          取消
        </Button>,
        <Button key="create" type="primary" onClick={onFinish}>
          确定
        </Button>,
      ]}
      onCancel={handleCancel}
    >
      <Form
        form={form}
        initialValues={formInitialValues}
        layout="vertical"
      >
        <Form.Item name="name" label="名称">
          <Input
            placeholder="请输入"
            className={styles.inputConfigPath}
          />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea
            placeholder="这里是描述，可能很长也可能很短的一段话..."
            rows={4}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default WorkscpaceForm;
