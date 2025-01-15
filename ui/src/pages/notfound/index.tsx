import React from 'react'
import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'

import styles from './styles.module.less'

const NotFound = () => {
  const navigate = useNavigate()
  function goBack() {
    navigate('/search')
  }

  return (
    <div className={styles.container}>
      <Result
        status="404"
        title="404"
        subTitle="Not Found Page"
        extra={
          <Button type="primary" onClick={goBack}>
            Back to home page
          </Button>
        }
      />
    </div>
  )
}

export default NotFound
