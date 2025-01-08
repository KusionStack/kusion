import React, { memo } from 'react'
import {
  CodeOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons'

import styles from './style.module.less'

const iconStyle = { marginRight: 5 }


const NavRight = ({ onClick, selectedKey }) => {

  function handleClick() {
    onClick('/backends')
  }
  return (
    <div
      className={styles.nav_right}
    >
      <div className={styles.nav_right_item} onClick={handleClick}
        style={{ borderBottom: selectedKey === '/backends' ? '2px solid #fff' : 'none' }}
      >
        <SettingOutlined style={iconStyle} />Backends
      </div>
      <div className={styles.nav_right_item}>
        <span
          onClick={() => {
            window.open('https://www.kusionstack.io/karpor')
          }}
        >
          <CodeOutlined style={iconStyle} />
          Document
        </span>
      </div>
      <div className={styles.nav_right_item}>
        <span
          onClick={() => {
            window.open('https://www.kusionstack.io/karpor')
          }}
        >
          <QuestionCircleOutlined style={iconStyle} />
          Help&Fallback
        </span>
      </div>
    </div>
  )
}

export default memo(NavRight)
