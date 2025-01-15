import React, { useState } from 'react'
import { Button, Drawer, message, Space } from 'antd'
import YamlEditor from '@/components/yamlEditor'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk'
import { yaml2json } from '@/utils/tools'

const EditYamlDrawer = ({ yamlData, open, handleSubmit, handleClose }) => {
  const [yamlStr, setYamlStr] = useState(yamlData)

  function handleChange(val) {
    setYamlStr(val)
  }

  function handleCancel() {
    setYamlStr(yamlData)
    handleClose()
  }

  function onSubmit() {
    handleSubmit(yamlStr)
  }

  async function validateYaml() {
    const response: any = await WorkspaceService.validateWorkspaceConfigs({
      body: yamlStr ? yaml2json(yamlStr)?.data : {}
    })
    if (response?.data?.success) {
      message.success('Validate Successful')
    } else {
      message.error(response?.data?.message)
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