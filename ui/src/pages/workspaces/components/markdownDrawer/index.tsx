import React, { useState } from 'react'
import { Button, Drawer, Space } from 'antd'
import Markdown from 'react-markdown'


const MarkdownDrawer = ({ yamlData, open, handleSubmit, handleClose, validateYaml }) => {

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
      title={'Generate kcl.mod'}
      open={open}
      onClose={handleCancel}
      width='80%'
    >
      <Markdown>{yamlStr}</Markdown>
    </Drawer>
  )

}

export default MarkdownDrawer