import React, { useEffect, useState } from 'react';
import { Modal, Button, Form, Select, Input, message, Collapse, theme, Radio, Switch } from 'antd';

import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import { CaretRightOutlined } from '@ant-design/icons';

import styles from './styles.module.less';

const RunsForm = ({ open, handleClose, handleSubmit, runsTypes }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const { token } = theme.useToken();
  const [workspaceList, setWorkspaceList] = useState([])
  const formInitialValues = {
    name: '',
    description: '',
    projectSource: '',
    configPath: '',
    organization: '',
  };

  function handleCancel() {
    form.resetFields();
    handleClose()
  }

  async function getWorkspaceList() {
    try {
      const response: any = await WorkspaceService.listWorkspace();
      if (response?.data?.success) {
        setWorkspaceList(response?.data?.data?.workspaces);
      } else {
        message.error(response?.data?.messaage)
      }
    } catch (error) {

    }

  }


  useEffect(() => {
    getWorkspaceList()
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
      message.error('提交失败');
    } finally {
      setLoading(false);
    }
  };

  const getItems = (panelStyle) => [
    {
      key: 'advancedSetting',
      label: 'Advanced Setting',
      children: <div>
        <p>Import Existing Resource</p>
        <Form.Item name="isExecting">
          <Switch />
        </Form.Item>
        {
          form.getFieldValue('isExecting') &&
          <Form.Item name="resources" label="resource map">
            <Input.TextArea />
          </Form.Item>
        }
      </div>,
      style: panelStyle,
    }
  ];

  const panelStyle: React.CSSProperties = {
    marginBottom: 24,
    background: '#fff',
    borderRadius: token.borderRadiusLG,
    border: 'none',
  };

  return (
    <Modal
      open={open}
      title="Create Runs"
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
        <Form.Item name="type" label="Type">
          <Select
            placeholder="Please Select Workspace"
            className={styles.selectInput}
          >
            {
              Object.entries(runsTypes)?.map(([key, value]: any) => {
                return <Select.Option key={key} value={key}>{value}</Select.Option>
              })
            }
          </Select>
        </Form.Item>
        <Form.Item name="workspace" label="Workspace">
          <Select
            placeholder="Please Select Workspace"
            className={styles.selectInput}
          >
            {
              workspaceList?.map(item => {
                return <Select.Option key={item?.id} value={item?.id}>{item?.name}</Select.Option>
              })
            }
          </Select>
        </Form.Item>
        <Form.Item noStyle>
          <Collapse
            bordered={false}
            defaultActiveKey={['1']}
            expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} />}
            style={{ background: token.colorBgContainer, marginLeft: -16 }}
            items={getItems(panelStyle)}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default RunsForm;
