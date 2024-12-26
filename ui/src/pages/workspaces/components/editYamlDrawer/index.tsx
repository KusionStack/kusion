import React, { useState } from 'react'
import { Button, Drawer, message, Space } from 'antd'
import YamlEditor from '@/components/yamlEditor'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk'
import { josn2yaml } from '@/utils/tools'

const EditYamlDrawer = ({ yamlData, open, handleSubmit, handleClose }) => {
  console.log(yamlData, "====EditYamlDrawer yamlData====")
  const yamlSs = josn2yaml(yamlData)?.data
  const [yamlStr, setYamlStr] = useState(yamlSs)

  function handleChange(val) {
    console.log(val, "====handleChange====")
    setYamlStr(val)
  }

  function handleCancel() {
    setYamlStr(yamlSs)
    handleClose()
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


  console.log(yamlStr, "===yamlStr====")
  
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
      {
        yamlStr &&
        <YamlEditor value={yamlStr} readOnly={false} onChange={handleChange} themeMode='DARK' />
      }
    </Drawer>
  )

}

export default EditYamlDrawer