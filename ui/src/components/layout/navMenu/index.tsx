import React, { memo } from 'react'
import { Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
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


const LayoutPage = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const { isLogin, isUnsafeMode } = useSelector((state: any) => state.globalSlice)

  function handleMenuClick(event) {
    if (event.key === '/search') {
      navigate('/search')
    } else if (!isLogin && !isUnsafeMode && ['/login']?.includes(pathname)) {
      return
    } else if (event?.domEvent.metaKey && event?.domEvent.button === 0) {
      const { origin } = window.location
      window.open(`${origin}${event.key}`)
    } else {
      navigate(event.key)
    }
  }

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


  return (
    <div className={styles.nav_menu}>
      <Menu
        style={{ border: 'none' }}
        mode="horizontal"
        // theme='dark'
        selectedKeys={[pathname]}
        items={getMenuItems()}
        onClick={handleMenuClick}
      />
    </div>
  )
}

export default memo(LayoutPage)
