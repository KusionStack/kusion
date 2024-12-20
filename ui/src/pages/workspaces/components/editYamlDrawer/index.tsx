import React, { useState } from 'react'
import { Button, Drawer, Space } from 'antd'
import YamlEditor from '@/components/yamlEditor'

const EditYamlDrawer = ({ yamlData, open, handleSubmit, handleClose, validateYaml }) => {

  const [yamlStr, setYamlStr] = useState(yamlData)

  function handleChange(val) {
    console.log(val, "====handleChange====")
    setYamlStr(val)
  }

  function handleCancel() {
    setYamlStr(yamlData)
  }

  return (
    <Drawer
      title={'YAML'}
      open={open}
      onClose={handleCancel}
      width='80%'
      extra={
        <Space>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={validateYaml}>Validate</Button>
          <Button type='primary' onClick={handleSubmit}>Submit</Button>
        </Space>
      }
    >
      <YamlEditor value={yamlStr} readOnly={false} onChange={handleChange} themeMode='DARK' />
    </Drawer>
  )

}

export default EditYamlDrawer