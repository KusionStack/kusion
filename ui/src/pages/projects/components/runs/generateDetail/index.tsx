import React from 'react'
import { Drawer } from 'antd'
import YamlEditor from '@/components/yamlEditor'
import { josn2yaml } from '@/utils/tools'

const GenerateDetail = ({ open, currentRecord, handleClose }) => {

  const yamlStr = josn2yaml(currentRecord?.result)

  return (
    <Drawer
      title={'Detail'}
      width="80%"
      open={open}
      onClose={handleClose}
    >
      <div style={{ height: '100%', overflowY: 'scroll' }}>
        <YamlEditor value={yamlStr?.data} readOnly={true} themeMode='DARK' />
      </div>
    </Drawer>
  )

}

export default GenerateDetail