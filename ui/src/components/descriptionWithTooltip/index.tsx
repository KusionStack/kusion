import React from 'react'
import { Tooltip } from 'antd'

import styles from './styles.module.less'

const DescriptionWithTooltip = ({ desc, width }) => {

  return (<Tooltip placement="topLeft" title={desc}>
    <div className={styles.descriptionWithTooltip} style={{ width }}>
      {desc}
    </div>
  </Tooltip>)
}

export default DescriptionWithTooltip