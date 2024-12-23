import React, { useState } from 'react'
import { Button, Drawer, Space } from 'antd'
import Markdown from 'react-markdown'


const MarkdownDrawer = ({ markdown, open, handleClose }) => {

  return (
    <Drawer
      title={'Generate kcl.mod'}
      open={open}
      onClose={handleClose}
      width='80%'
      extra={
        [
          <Button onClick={handleClose}>Close</Button>
        ]
      }
    >
      <div style={{ background: '#000', color: '#fff', padding: 20, height: '100%', overflowY: 'auto' }}>
        <Markdown>{markdown}</Markdown>
      </div>
    </Drawer>
  )

}

export default MarkdownDrawer