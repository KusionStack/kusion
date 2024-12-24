import React, { useEffect, useRef, useState } from 'react'
import { Drawer, Segmented, Select } from 'antd'
import { default as AnsiUp } from 'ansi_up';
import CodeDiffView from '@/components/codeDiffView'

import styles from "./styles.module.less"

const PreviewDetail = ({ open, currentRecord, handleClose }) => {
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const ansi_up = new AnsiUp();
  const logRef = useRef<HTMLDivElement | null>(null);
  const { stepKeys, changeSteps } = currentRecord?.result && JSON.parse(currentRecord?.result)

  const [activeKey, setActiveKey] = useState('Exec Result')
  const [selectedResource, setSelectedResource] = useState(stepKeys?.[0])

  function handleChange(val) {
    setActiveKey(val)
  }


  function handleChangeResources(val) {
    setSelectedResource(val)
  }



  useEffect(() => {
    if (logRef && logRef.current) {
      const logHtml = ansi_up.ansi_to_html(currentRecord?.logs);
      logRef.current.innerHTML = logHtml;
    }
  }, [ansi_up, currentRecord?.logs, logRef]);

  return (
    <Drawer
      title={'Priview Detail'}
      open={open}
      width="80%"
      onClose={handleClose}
    >
      <div>
        <div style={{ marginBottom: 20 }}>
          <Segmented options={["Exec Result", "Exec Logs"]} value={activeKey} onChange={handleChange} />
        </div>
        {
          activeKey === 'Exec Result' && (
            <div>
              <div className={styles.results}>
                <Select value={selectedResource} style={{ width: 500, marginBottom: 10 }} onChange={handleChangeResources}>
                  {
                    stepKeys?.map(item => {
                      return <Select.Option key={item} value={item}>{item}</Select.Option>
                    })
                  }
                </Select>
              </div>
              <CodeDiffView oldContent={changeSteps?.[selectedResource]?.from ? JSON.stringify(changeSteps?.[selectedResource]?.from, null, 2) : ''} newContent={changeSteps?.[selectedResource]?.to ? JSON.stringify(changeSteps?.[selectedResource]?.to, null, 2) : ''} />
            </div>
          )
        }
        {
          activeKey === 'Exec Logs' && (
            <div style={{ background: '#000', color: '#fff', padding: 20 }}>
              <div ref={logRef}></div>
            </div>
          )
        }
      </div>
    </Drawer>
  )

}

export default PreviewDetail