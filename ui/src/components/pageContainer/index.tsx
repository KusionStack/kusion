import React from 'react'
import styles from './style.module.less'
import { Card } from 'antd'

const PageContainer = ({ title, children }) => {
  return (
    <div className={styles.page_container}>
      <div className={styles.page_container_header}>
        {typeof title === 'string' ? <h2>{title}</h2> : title}
      </div>
      <div className={styles.page_container_content}>
        <Card>{children}</Card>
      </div>
    </div>
  )
}

export default PageContainer
