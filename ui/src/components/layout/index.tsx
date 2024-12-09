import React, { memo, useEffect } from 'react'
import { Menu } from 'antd'
import {
  ClusterOutlined,
  FundOutlined,
  SearchOutlined,
  LeftOutlined,
  RightOutlined,
  QuestionOutlined,
  FileFilled,
  QuestionCircleFilled,
  SmileFilled,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import {
  setServerConfigMode,
  setVersionNumber,
  setGithubBadge,
  setIsUnsafeMode,
  setCollapsed,
} from '@/store/modules/globalSlice'
import logo from '@/assets/img/logo.svg'
import { useAxios } from '@/utils/request'

import styles from './style.module.less'

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
  const dispatch = useDispatch()
  const { isLogin, isUnsafeMode, collapsed } = useSelector(
    (state: any) => state.globalSlice,
  )
  // const [collapsed, setCollapsed] = useState(false)

  const toggleCollapsed = () => {
    dispatch(setCollapsed(!collapsed))
  }

  const { response } = useAxios({
    url: '/server-configs',
    option: { params: {} },
    manual: false,
    method: 'GET',
  })

  useEffect(() => {
    if (response) {
      dispatch(setServerConfigMode(response?.CoreOptions?.ReadOnlyMode))
      dispatch(setIsUnsafeMode(!response?.CoreOptions?.EnableRBAC))
      dispatch(setVersionNumber(response?.Version))
      dispatch(setGithubBadge(response?.CoreOptions?.GithubBadge))
    }
  }, [response, dispatch])

  const menuItems = [
    getItem('Projects', '/projects', <SearchOutlined />),
    getItem('Workspaces', '/workspaces', <FundOutlined />),
    getItem('Modules', '/modules', <ClusterOutlined />),
    getItem('Insights', '/insights', <ClusterOutlined />),
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

  useEffect(() => {
    if (
      !isLogin &&
      !isUnsafeMode &&
      !['/login', '/', '/search']?.includes(pathname)
    ) {
      navigate('/login')
    }
  }, [isLogin, isUnsafeMode, navigate, pathname])

  const toogle_slide_style = {
    width: 10,
    height: 10,
  }

  const bottom_icon_style = {
    color: '#646566'
  }

  return (
    <div className={styles.wrapper}>
      <div
        className={styles.layout_left}
        style={{ width: collapsed ? '' : 200 }}
      >
        <div
          className={styles.toggle_collapsed_sider}
          onClick={toggleCollapsed}
          style={{ left: !collapsed ? 199 : 79 }}
        >
          {!collapsed ? (
            <LeftOutlined style={toogle_slide_style} />
          ) : (
            <RightOutlined style={toogle_slide_style} />
          )}
        </div>
        <div
          className={styles.header}
          style={{ justifyContent: collapsed ? 'center' : 'flex-start' }}
          onClick={() => navigate('/')}
        >
          <div className={styles.header_left}>
            <div className={styles.header_left_logo}>
              <img src={logo} alt="logo" />
            </div>
            {!collapsed && <h4 className={styles.title}>Kusion</h4>}
          </div>
        </div>
        <Menu
          style={{ height: '100%', border: 'none' }}
          mode="vertical"
          inlineCollapsed={collapsed}
          selectedKeys={[pathname]}
          items={getMenuItems()}
          onClick={handleMenuClick}
        />
        <div className={styles.bottom}>
          <div className={styles.bottom_item} style={{ justifyContent: collapsed ? 'center' : 'flex-start' }}>
            <div className={styles.bottom_item_icon}>
              <FileFilled style={bottom_icon_style} />
            </div>
            {!collapsed && <span className={styles.bottom_item_text}>Document</span>}
          </div>
          <div className={styles.bottom_item} style={{ justifyContent: collapsed ? 'center' : 'flex-start' }}>
            <div className={styles.bottom_item_icon}>
              <QuestionCircleFilled style={bottom_icon_style} />
            </div>
            {!collapsed && <span className={styles.bottom_item_text}>Help&Feelback</span>}
          </div>
          <div className={styles.bottom_item} style={{ justifyContent: collapsed ? 'center' : 'flex-start' }}>
            <div className={styles.bottom_item_icon}>
              <SmileFilled style={bottom_icon_style} />
            </div>
            {!collapsed && <span className={styles.bottom_item_text}>User</span>}
          </div>
        </div>
      </div>
      <div className={styles.right}>
        {!isLogin &&
          !isUnsafeMode &&
          !['/login', '/', '/search']?.includes(pathname) ? null : (
          <div className={styles.right_content}>
            <Outlet />
          </div>
        )}
      </div>
    </div>
  )
}

export default memo(LayoutPage)
