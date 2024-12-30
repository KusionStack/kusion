import React from 'react'
import { SmileOutlined } from '@ant-design/icons';
import { Button, Card, Result } from 'antd';
import { useNavigate } from 'react-router-dom';

import styles from "./styles.module.less"

const Insights = () => {
  const navigate = useNavigate()

  function goHome() {
    navigate('/projects')
  }
  return (
    <div className={styles.insight_container}>
      <Card style={{ width: '100%', height: '100%' }}>
        <Result
          icon={<SmileOutlined />}
          title="Coming soon..."
          extra={<Button type="primary" onClick={goHome}>Back Home</Button>}
        />
      </Card>
    </div>
  )
}

export default Insights
