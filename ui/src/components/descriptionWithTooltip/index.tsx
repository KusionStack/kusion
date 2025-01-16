import React from 'react'
import { Tooltip } from 'antd'

import styles from './styles.module.less'

type IProps = {
  desc: string
  width?: number | string
}

const DescriptionWithTooltip = ({ desc, width }: IProps) => {

  return (<Tooltip placement="topLeft" title={desc}>
    <div className={styles.descriptionWithTooltip} style={{ width }}>
      {desc}
    </div>
  </Tooltip>)
}

export default DescriptionWithTooltip