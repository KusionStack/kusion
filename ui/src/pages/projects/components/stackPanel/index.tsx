import React, { useEffect, useRef, useState } from 'react'
import styles from "./styles.module.less"
import { Button, Card, Col, DatePicker, Form, Input, Row, Space, Table, Tabs, Tag } from 'antd'
import { ArrowLeftOutlined, CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { ProjectService } from '@kusionstack/kusion-api-client-sdk'
import Runs from '../runs'
import ResourceGraph from '../resourceGraph'

type TypeSearchParams = {
  name?: string
  org?: string
  creator?: string
  createTime?: string
} | undefined

const initialItems = [
  { label: 'Resource Graph', key: 'ResourceGraph' },
  { label: 'Runs', key: 'Runs' },
];


type TargetKey = React.MouseEvent | React.KeyboardEvent | string;

const StackPanel = ({ stackName }) => {
  const [form] = Form.useForm();
  const [searchParams, setSearchParams] = useState<TypeSearchParams>();
  const [dataSource, setDataSource] = useState([])
  const [open, setOpen] = useState<boolean>(false);
  const [activeKey, setActiveKey] = useState(initialItems?.[0]?.key);
  const [items, setItems] = useState(initialItems);
  const newTabIndex = useRef(0);

  function onChange(newActiveKey: string) {
    setActiveKey(newActiveKey);
  };


  return (
    <div>
      <Tabs
        onChange={onChange}
        activeKey={activeKey}
        items={items}
      />
      {activeKey === 'Runs' && <Runs />}
      {activeKey === 'ResourceGraph' && <ResourceGraph />}
    </div>
  )
}

export default StackPanel
