import React, { useEffect, useState } from 'react';
import { Modal, Button, Form, Select, Input, message, Collapse, theme, Radio } from 'antd';

import styles from './styles.module.less';
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import { CaretRightOutlined } from '@ant-design/icons';

const RunsForm = ({ open, handleClose, handleSubmit }: any) => {
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
      console.log(response, "====getWorkspaceList response===")
      if (response?.data?.success) {
        setWorkspaceList(response?.data?.data);
      } else {
        message.error("请求失败")
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
    console.log(form.getFieldsValue(), "==sdada======")
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

  const getItems = (panelStyle) => [
    {
      key: 'advancedSetting',
      label: 'Advanced Setting',
      children: <div>
        <p>Import Existing Resource</p>
        <Form.Item name="isExecting">
          <Radio.Group>
            <Radio value="YES">YES</Radio>
            <Radio value="NO">NO</Radio>
          </Radio.Group>
        </Form.Item>
        <Form.Item name="resources" label="resource map">
          <Input.TextArea />
        </Form.Item>
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
          取消
        </Button>,
        <Button htmlType='submit' key="create" type="primary" onClick={onFinish}>
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
        <Form.Item name="type" label="Type">
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
