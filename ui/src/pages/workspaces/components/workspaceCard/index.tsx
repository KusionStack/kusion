import React from "react";
import { Button, Popconfirm, Tooltip } from "antd";
import workspaceSvg from "@/assets/img/workspace.svg"
import {
  DeleteOutlined,
  EditOutlined,
} from '@ant-design/icons'
import styles from "./styles.module.less"

const WorkspaceCard = ({ title, desc, nickName, createDate, onClick, onDelete, handleEdit }) => {

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
            <Button type='link' onClick={() => handleEdit({
              name: title,
              description: desc,
            })}>
              <EditOutlined />
            </Button>
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
              <Button type='link' danger><DeleteOutlined /></Button>
            </Popconfirm>
          </div>
        </div>
        <div className={styles.workspace_card_content} onClick={onClick}>
          <Tooltip title={desc} placement="topLeft">
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