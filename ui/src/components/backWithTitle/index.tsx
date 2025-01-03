import React from "react";

import styles from "./styles.module.less"
import { ArrowLeftOutlined } from "@ant-design/icons";

const BackWithTitle = ({ title, handleBack }) => {
  return (
    <div className={styles.kusion_back}>
      <div className={styles.kusion_back_arrow} onClick={handleBack}>
        <ArrowLeftOutlined style={{ fontSize: 20 }} />
      </div>
      <h3>{title}</h3>
    </div>
  )
}

export default BackWithTitle