import React, { useEffect, useRef, useState } from 'react'
import { Badge, Drawer, Segmented, Select, Tag } from 'antd'
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
      logRef.current.style.whiteSpace = 'pre-wrap';
    }
  }, [ansi_up, currentRecord?.logs, logRef]);

  const dotStyle = {
    background: changeSteps?.[selectedResource]?.action === 'Undefined' ? '#ff4d4f' : changeSteps?.[selectedResource]?.action === 'UnChanged' ? "rgba(0,0,0,0.25)" : '#faad14'
  }

  return (
    <Drawer
      title={'Preview Detail'}
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
            <>
              <div className={styles.prrviewContainer}>
                <Select value={selectedResource} style={{ width: 500, marginBottom: 10 }} onChange={handleChangeResources}>
                  {
                    stepKeys?.map(item => {
                      return <Select.Option key={item} value={item}>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <div style={{ flex: 1, overflowX: 'hidden', textOverflow: 'ellipsis' }}>{item}</div>
                          <div>
                            <Tag color={
                              changeSteps?.[item]?.action === 'Undefined' ? 'error' : changeSteps?.[item]?.action === 'UnChanged' ? "default" : 'warning'
                            } >
                              {changeSteps?.[item]?.action}
                            </Tag>
                          </div>
                        </div>
                      </Select.Option>
                    })
                  }
                </Select>
                {
                  selectedResource && <div className={styles.status}>
                    <div className={styles.animate_wave}>
                      <div
                        style={dotStyle}
                        className={`${styles.animate_circle} ${styles.animate_inner}`}>
                      </div>
                      <div
                        style={dotStyle}
                        className={`${styles.animate_circle} ${styles.animate_middle}`}>
                      </div>
                      <div
                        style={dotStyle}
                        className={`${styles.animate_circle} ${styles.animate_outer}`}>
                      </div>
                    </div>
                    {changeSteps?.[selectedResource]?.action}
                  </div>
                }
              </div>
              <CodeDiffView oldContent={changeSteps?.[selectedResource]?.from ? JSON.stringify(changeSteps?.[selectedResource]?.from, null, 2) : ''} newContent={changeSteps?.[selectedResource]?.to ? JSON.stringify(changeSteps?.[selectedResource]?.to, null, 2) : ''} />
            </>
          )
        }
        {
          activeKey === 'Exec Logs' && (
            <div style={{ background: '#000', color: '#fff', padding: 20 }}>
              <div ref={logRef} />
            </div>
          )
        }
      </div>
    </Drawer >
  )

}

export default PreviewDetail