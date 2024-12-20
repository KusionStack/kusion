import React from "react";

import styles from "./styles.module.less"
import { PlusOutlined } from "@ant-design/icons";
import { Tooltip } from "antd";

const WorkspaceCard = ({ title, desc, nickName, createDate, onClick }) => {

  return (
    <div className={styles.workspace_card} onClick={onClick}>
      <div className={styles.workspace_card_container}>
        <div className={styles.workspace_card_header}>
          <div className={styles.workspace_card_icon}><PlusOutlined /></div>
          <div className={styles.workspace_card_title}>{title}</div>
        </div>
        <div className={styles.workspace_card_content}>
          <Tooltip title={desc}>
            <div className={styles.kusion_card_content_desc}>
              {desc}
            </div>
          </Tooltip>
        </div>
        <div className={styles.workspace_card_footer}>
          <div className={styles.workspace_card_footer_left}>
            {nickName}
          </div>
          <div className={styles.workspace_card_footer_right}>
            {createDate}
          </div>
        </div>
      </div>
    </div>
  )
}

export default WorkspaceCard;