import React, { useEffect, useState } from 'react';
import { Modal, Button, Form, Select, Input, message, Collapse, theme, Radio, Switch, Space } from 'antd';

import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import { CaretRightOutlined, MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

import styles from './styles.module.less';

const RunsForm = ({ open, handleClose, handleSubmit, runsTypes }: any) => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const { token } = theme.useToken();
  const [workspaceList, setWorkspaceList] = useState([])
  const [switchEnable, setSwitchEnable] = useState<boolean>(false);

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

  // Submit form
  const onFinish = async () => {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      const values = form.getFieldsValue();
      handleSubmit(values)
    } catch (e) {
      message.error('Submit failed');
    } finally {
      setLoading(false);
    }
  };

  function handleChangeSwitch(checked) {
    setSwitchEnable(checked)
  }

  const getItems = (panelStyle) => [
    {
      key: 'advancedSetting',
      label: 'Advanced Setting',
      children: <div>
        <p>Import Existing Resource</p>
        <Form.Item name="isExecting">
          <Switch onChange={handleChangeSwitch} />
        </Form.Item>
        {
          switchEnable &&
          <Form.List name="configs">
            {(fields, { add, remove }) => {
              return (
                <>
                  {fields.map(({ key, name, ...restField }, index) => {
                    return (
                      <Form.Item
                        label={index === 0 ? 'Resources' : ''}
                        required={true}
                        key={key}
                        style={{ marginBottom: 0 }}
                      >
                        <Space
                          style={{ display: 'flex', marginBottom: 8 }}
                          align="baseline"
                        >
                          <Form.Item
                            {...restField}
                            name={[name, 'key']}
                            rules={[
                              {
                                required: false,
                                message: 'Missing first name',
                              },
                            ]}
                          >
                            <Input
                              style={{ width: 280 }}
                            />
                          </Form.Item>
                          <Form.Item
                            {...restField}
                            name={[name, 'value']}
                            rules={[{ required: false, message: 'Required' }]}
                          >
                            <Input
                              style={{ width: 280 }}
                            />
                          </Form.Item>
                          {fields?.length > 1 && (
                            <MinusCircleOutlined onClick={() => remove(name)} />
                          )}
                        </Space>
                      </Form.Item>
                    )
                  })}
                  <Form.Item>
                    <Button
                      type="dashed"
                      onClick={() => add()}
                      block
                      icon={<PlusOutlined />}
                    >
                      Add
                    </Button>
                  </Form.Item>
                </>
              )
            }}
          </Form.List>
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
      width={650}
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
        initialValues={{
          configs: [
            {
              key: '',
              value: '',
            },
          ],
        }}
        layout="vertical"
      >
        <Form.Item name="type" label="Type">
          <Select
            placeholder="Please Select Type"
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
                return <Select.Option key={item?.id} value={item?.name}>{item?.name}</Select.Option>
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
