import React, { memo } from 'react'
import {
  CodeOutlined,
  QuestionCircleOutlined,
  UserOutlined,
} from '@ant-design/icons'

import styles from './style.module.less'

const iconStyle = { marginRight: 5 }

const NavRight = () => {
  return (
    <div
      className={styles.nav_right}
    >
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
      <div className={styles.nav_right_item}>
        <span
          onClick={() => {
            window.open('https://www.kusionstack.io/karpor')
          }}
        >
          <UserOutlined style={iconStyle} />
          Role
        </span>
      </div>
    </div>
  )
}

export default memo(NavRight)
