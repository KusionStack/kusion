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
        <div className={styles.workspace_card_header} style={{cursor: 'pointer'}} onClick={onClick}>
          <div className={styles.workspace_card_header_left}>
            <div className={styles.workspace_card_icon}>
              <img src={workspaceSvg} alt="svgIcon" />
            </div>
            <div className={styles.workspace_card_title}>{title}</div>
          </div>
          <div>
            <Button type='link' onClick={(e) => {
              e.stopPropagation();
              handleEdit({
                name: title,
                description: desc,
              });
            }}>
              <EditOutlined />
            </Button>
            <Popconfirm
              title={<span style={{fontSize: '18px'}}>Delete the workspace</span>}
              description={<span style={{fontSize: '16px'}}>Are you sure to delete this workspace?</span>}
              onConfirm={(event) => {
                event.stopPropagation()
                onDelete()
              }}
              onCancel={(e) => {
                e.stopPropagation()
              }}
              okText={<span style={{fontSize: '16px'}}>Yes</span>}
              cancelText={<span style={{fontSize: '16px'}}>No</span>}
              overlayStyle={{
                width: '330px'
              }}
              okButtonProps={{
                style: {height: '30px', width: '60px'}
              }}
              cancelButtonProps={{
                style: {height: '30px', width: '60px'}
              }}
            >
              <Button type='link' danger onClick={(e) => e.stopPropagation()}><DeleteOutlined /></Button>
            </Popconfirm>
          </div>
        </div>
        <div className={styles.workspace_card_content} onClick={onClick} style={{cursor: 'pointer'}}>
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