import React from 'react'
import { Button, Drawer } from 'antd'
import MarkdownView from '@/components/markdownView'


const MarkdownDrawer = ({ markdown, open, handleClose }) => {

  return (
    <Drawer
      key="markdown"
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
        <MarkdownView markdown={markdown} />
      </div>
    </Drawer>
  )

}

export default MarkdownDrawer