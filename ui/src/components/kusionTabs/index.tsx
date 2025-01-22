import React from 'react'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import classNames from 'classnames'
import { Tooltip } from 'antd'

import styles from './styles.module.less'

export const KusionTabs = ({
  items,
  addIsDiasble,
  activeKey,
  handleClickItem,
  onEdit,
  disabledAdd,
}) => {
  function handleActionIcon(event, id) {
    event.preventDefault()
    event.stopPropagation()
    onEdit('edit', id)
  }
  function handleAdd() {
    if (disabledAdd) return
    onEdit('add')
  }
  function handleChangeTab(key) {
    if (activeKey === key) return
    handleClickItem(key)
  }
  return (
    <div className={styles.tabs_wrapper}>
      <div className={styles.tabs_container}>
        <div className={styles.tabs}>
          {items?.map(item => {
            return (
              <div
                key={item?.id}
                className={classNames(styles.tab, {
                  [styles.active_tab]: item?.id === activeKey,
                })}
                onClick={() => handleChangeTab(item?.id)}
              >
                <div className={styles.label}>{item?.label}</div>
                <div
                  className={styles.edit_icon}
                  onClick={event => handleActionIcon(event, item?.id)}
                >
                  <CloseOutlined />
                </div>
              </div>
            )
          })}
        </div>
        <Tooltip title="New Stack">
          <div className={styles.add_box} onClick={handleAdd}>
            <PlusOutlined style={{ fontSize: 14, color: '#fff' }} />
          </div>
        </Tooltip>
      </div>
    </div>
  )
}
