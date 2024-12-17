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

  return (
    <Modal
      open={open}
      title="创建新任务"
      footer={[
        <Button key="cancel" onClick={() => { }}>
          取消
        </Button>,
        <Button key="create" type="primary" onClick={onFinish}>
          确定
        </Button>,
      ]}
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
        <Form.Item name="projectSource" label="项目来源">
          <Select
            placeholder="请选择 projects source"
            className={styles.selectInput}
          />
        </Form.Item>
        <Form.Item name="configPath" label="配置路径">
          <Input
            placeholder="请输入 config path in source"
            className={styles.inputConfigPath}
          />
        </Form.Item>
        <Form.Item name="organization" label="所属组织">
          <Select placeholder="请选择所属组织" className={styles.selectInput} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ProjectForm;
