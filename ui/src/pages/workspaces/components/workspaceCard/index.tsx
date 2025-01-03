import React from "react";
import { Tooltip } from "antd";
import workspaceSvg from "@/assets/img/workspace.svg"

import styles from "./styles.module.less"

const WorkspaceCard = ({ title, desc, nickName, createDate, onClick }) => {

  return (
    <div className={styles.workspace_card} onClick={onClick}>
      <div className={styles.workspace_card_container}>
        <div className={styles.workspace_card_header}>
          <div className={styles.workspace_card_icon}>
            <img src={workspaceSvg} alt="svgIcon" />
          </div>
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