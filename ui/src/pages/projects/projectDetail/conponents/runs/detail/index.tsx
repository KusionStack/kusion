import React, { useState } from 'react'

import styles from "./styles.module.less"
import { Drawer, Segmented } from 'antd'
import CodeDiffView from '@/components/codeDiffView'
import { mockYaml, mockNewYaml } from '@/utils/tools'
import Markdown from 'react-markdown'

const RunsDetail = ({ title }) => {

  const [activeKey, setActiveKey] = useState('EXEC Result')

  function handleChange(val) {
    setActiveKey(val)
  }

  return (
    <Drawer
      title={title}

    >
      <div>
        <Segmented options={["EXEC Result", "EXEC Logs"]} value={activeKey} onChange={handleChange} />
        {
          activeKey === 'EXEC Result' && (
            <div>
              <CodeDiffView oldContent={mockYaml} newContent={mockNewYaml} />
            </div>
          )
        }
        {
          activeKey === 'EXEC Logs' && (
            <div>
              <Markdown>{""}</Markdown>
            </div>
          )

        }

      </div>
    </Drawer>
  )

}