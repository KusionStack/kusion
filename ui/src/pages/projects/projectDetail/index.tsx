import React, { useEffect, useRef, useState } from 'react'
import { Card, Form, message, Tabs, } from 'antd'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons'
import { StackService } from '@kusionstack/kusion-api-client-sdk'
import StackPanel from "../components/stackPanel"
import BackWithTitle from '@/components/backWithTitle'
import StackForm from '../components/stackForm'

import styles from "./styles.module.less"

type TargetKey = React.MouseEvent | React.KeyboardEvent | string;

const ProjectDetail = () => {
  const navigate = useNavigate()
  const urlPrams = useParams()
  const [activeKey, setActiveKey] = useState('');
  const [items, setItems] = useState([]);
  const [stackFormOpen, setStackFormOpen] = useState(false)

  function onChange(newActiveKey: string) {
    setActiveKey(newActiveKey);
  };


  const add = () => {
    setStackFormOpen(true)
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

  async function handleSubmit(values) {
    const response: any = await StackService.createStack({
      body: {
        ...values,
        projectID: urlPrams?.projectId
      }
    })
    if (response?.data?.success) {
      message.success('Create Successful')
      handleClose()
      const newPanes = {
        label: values?.name,
        key: values?.name,
        closable: false
      }
      const newItems = [...items, newPanes]
      setItems(newItems)
    } else {
      message.error(response?.data?.message || 'Create Failed')
    }
  }
  function handleClose() {
    console.log("handleClose")
    setStackFormOpen(false)
  }

  async function getStackList(params) {
    try {
      const response: any = await StackService.listStack(params);
      if (response?.data?.success) {
        const resTabs = response?.data?.data?.map(item => {
          return {
            ...item,
            label: item?.name,
            key: item?.id,
          }
        })
        setItems(resTabs)
      } else {
        message.error(response?.data?.message)
      }
    } catch (error) {

    }
  }

  useEffect(() => {
    getStackList({})
  }, [])

  function handleBack() {
    navigate("/projects")
  }

  return (
    <div className={styles.project_detail}>
      <BackWithTitle title="项目名称" handleBack={handleBack} />
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
        <StackForm formData={{}} actionType="ADD" stackFormOpen={stackFormOpen} handleCancel={handleClose} handleSubmit={handleSubmit} />
      </Card>
    </div>
  )
}

export default ProjectDetail
