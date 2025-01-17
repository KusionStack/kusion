import React, { memo } from 'react'
import type { MenuProps } from 'antd'

import styles from "./style.module.less"

type MenuItem = Required<MenuProps>['items'][number]

function getItem(
  label: React.ReactNode,
  key: React.Key,
  icon?: React.ReactNode,
  children?: MenuItem[],
  type?: 'group',
  hidden?: boolean,
  disabled?: boolean,
): MenuItem {
  return {
    key,
    icon,
    children,
    label,
    type,
    hidden,
    disabled,
  } as MenuItem
}

const KusionMenu = ({ selectedKey, onClick }) => {

  const menuItems = [
    getItem('Projects', '/projects', null),
    getItem('Workspaces', '/workspaces', null),
    getItem('Modules', '/modules', null),
    getItem('Sources', '/sources', null),
    getItem('Insights', '/insights', null),
  ]

  function getMenuItems() {
    function loop(list) {
      return list
        ?.filter(item => !item?.hidden)
        ?.map(item => {
          if (item?.children) {
            item.children = loop(item?.children)
          }
          return item
        })
    }
    return loop(menuItems)
  }

  function handleKusionMenuClick(item) {
    onClick(item?.key)
  }


  return (
    <div className={styles.nav_menu}>
      <ul className={styles.kusion_menu_container}>
        {
          getMenuItems()?.map(item => {
            const isSeletced = item?.key === selectedKey
            return <li className={`${styles.kusion_menu_item} ${isSeletced ? styles.kusion_menu_item_acitve : ''}`} key={item?.key} onClick={() => handleKusionMenuClick(item)}>{item?.label}</li>
          })
        }
      </ul>

    </div>
  )
}

export default memo(KusionMenu)
