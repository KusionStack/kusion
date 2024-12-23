import React, { useState } from 'react'
import { Button, Drawer, message, Space } from 'antd'
import YamlEditor from '@/components/yamlEditor'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk'

const EditYamlDrawer = ({ yamlData, open, handleSubmit, handleClose }) => {

  const [yamlStr, setYamlStr] = useState(yamlData)

  function handleChange(val) {
    console.log(val, "====handleChange====")
    setYamlStr(val)
  }

  function handleCancel() {
    setYamlStr(yamlData)
  }

  function onSubmit() {
    handleSubmit(yamlStr)
  }

  async function validateYaml() {
    const response: any = await WorkspaceService.validateWorkspaceConfigs({
      body: yamlStr && JSON.parse(yamlStr || '{}')
    })
    if (response?.data?.success) {
      message.success('Validate Successful')
    } else {
      message.error(response?.data?.message || '请求失败')
    }
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
          <Button type='primary' onClick={onSubmit}>Submit</Button>
        </Space>
      }
    >
      <YamlEditor value={yamlStr} readOnly={false} onChange={handleChange} themeMode='DARK' />
    </Drawer>
  )

}

export default EditYamlDrawer