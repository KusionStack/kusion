import React, { useEffect, useRef, useState } from 'react'
import { Card, Form, Tabs, } from 'antd'
import { useNavigate } from 'react-router-dom'
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons'
import { StackService } from '@kusionstack/kusion-api-client-sdk'
import StackPanel from "./conponents/stackPanel"

import styles from "./styles.module.less"

type TargetKey = React.MouseEvent | React.KeyboardEvent | string;

const ProjectDetail = () => {
  const initialItems = [
    { label: 'Tab 1', key: '1', closable: false, },
    { label: 'Tab 2', key: '2', closable: false, },
  ];
  const navigate = useNavigate()
  const [form] = Form.useForm();
  const [activeKey, setActiveKey] = useState(initialItems[0].key);
  const [items, setItems] = useState(initialItems);
  const newTabIndex = useRef(0);

  function onChange(newActiveKey: string) {
    setActiveKey(newActiveKey);
  };


  const add = () => {
    const newActiveKey = `newTab${newTabIndex.current++}`;
    const newPanes = [...items];
    newPanes.push({ label: 'New Tab', key: newActiveKey, closable: false });
    setItems(newPanes);
    setActiveKey(newActiveKey);
  };

  const remove = (targetKey: TargetKey) => {
    let newActiveKey = activeKey;
    let lastIndex = -1;
    items.forEach((item, i) => {
      if (item.key === targetKey) {
        lastIndex = i - 1;
      }
    });
    const newPanes = items.filter((item) => item.key !== targetKey);
    if (newPanes.length && newActiveKey === targetKey) {
      if (lastIndex >= 0) {
        newActiveKey = newPanes[lastIndex].key;
      } else {
        newActiveKey = newPanes[0].key;
      }
    }
    setItems(newPanes);
    setActiveKey(newActiveKey);
  };

  const onEdit = (
    targetKey: React.MouseEvent | React.KeyboardEvent | string,
    action: 'add' | 'remove',
  ) => {
    if (action === 'add') {
      add();
    } else {
      remove(targetKey);
    }
  };

  function handleSubmit(values) {
    console.log(values, "handleSubmit")
  }
  function handleClose() {
    console.log("handleClose")
  }
  function handleCreate() {
    console.log("=====handleCreate=====")
  }
  function handleReset() {
    form.resetFields();
  }
  function handleSearch() {
    const values = form.getFieldsValue()
    console.log(values, "=====handleSearch=====")
  }

  function handleClear(key) {
    form.setFieldValue(key, undefined)
    handleSearch()
  }

  async function getList(params) {
    try {
      const response: any = await StackService.listStack(params);
      console.log(response?.data?.data, "======response====")
      if (response?.data?.success) {
        const res = response?.data?.data?.map(item => {
          return {
            ...item,
            label: item?.name,
            key: item?.id,
          }
        })
        setItems(res)
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    getList({})
  }, [])

  function handleBack() {
    navigate("/projects")
  }

  return (
    <div className={styles.project_detail}>
      <div className={styles.project_detail_action}>
        <div className={styles.project_detail_action_back} onClick={handleBack}><ArrowLeftOutlined style={{ fontSize: 20 }} /></div>
        <h3>项目名称</h3>
      </div>
      <Card>
        <Tabs
          style={{ border: 'none' }}
          type="editable-card"
          onChange={onChange}
          activeKey={activeKey}
          onEdit={onEdit}
          items={items}
          addIcon={(
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <PlusOutlined />
              <div style={{ width: 100 }}>Create Stack</div>
            </div>
          )}
        />
        <StackPanel stackName={activeKey} />
      </Card>
    </div>
  )
}

export default ProjectDetail
