import React from "react";
import { Button, Popconfirm, Tooltip } from "antd";
import workspaceSvg from "@/assets/img/workspace.svg"

import styles from "./styles.module.less"

const WorkspaceCard = ({ title, desc, nickName, createDate, onClick, onDelete }) => {

  return (
    <div className={styles.workspace_card}>
      <div className={styles.workspace_card_container}>
        <div className={styles.workspace_card_header}>
          <div className={styles.workspace_card_header_left}>
            <div className={styles.workspace_card_icon}>
              <img src={workspaceSvg} alt="svgIcon" />
            </div>
            <div className={styles.workspace_card_title}>{title}</div>
          </div>
          <div>
            <Popconfirm
              title="Delete the workspace"
              description="Are you sure to delete this workspace?"
              onConfirm={(event) => {
                event.stopPropagation()
                onDelete()
              }}
              okText="Yes"
              cancelText="No"
            >
              <Button type='link' danger>Delete</Button>
            </Popconfirm>
          </div>
        </div>
        <div className={styles.workspace_card_content} onClick={onClick}>
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