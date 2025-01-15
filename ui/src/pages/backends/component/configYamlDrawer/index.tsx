import React from 'react'
import { Button, Drawer, Space } from 'antd'
import YamlEditor from '@/components/yamlEditor'

const ConfigYamlDrawer = ({ yamlData, open, handleClose }) => {

  return (
    <Drawer
      title={'YAML'}
      open={open}
      onClose={handleClose}
      width='80%'
      extra={
        <Space>
          <Button onClick={handleClose}>Close</Button>
        </Space>
      }
    >
      <YamlEditor value={yamlData} readOnly={false} themeMode='DARK' />
    </Drawer>
  )

}

export default ConfigYamlDrawer