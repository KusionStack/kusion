import React, { useState } from 'react'
import { Tabs } from 'antd'
import Runs from '../runs'
import ResourceGraph from '../resourceGraph'

const tabsItems = [
  { label: 'Resource Graph', key: 'ResourceGraph' },
  { label: 'Runs', key: 'Runs' },
];


const StackPanel = ({ stackId }) => {
  const [activeKey, setActiveKey] = useState(tabsItems?.[0]?.key);

  function onChange(newActiveKey: string) {
    setActiveKey(newActiveKey);
  };


  return (
    <div>
      <Tabs
        onChange={onChange}
        activeKey={activeKey}
        items={tabsItems}
      />
      {activeKey === 'Runs' && <Runs />}
      {activeKey === 'ResourceGraph' && <ResourceGraph stackId={stackId}/>}
    </div>
  )
}

export default StackPanel
