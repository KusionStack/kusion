import React, { memo, useEffect } from 'react'
import { Divider, Menu, Dropdown, Button, message } from 'antd'
import {
  ClusterOutlined,
  FundOutlined,
  QuestionCircleOutlined,
  SearchOutlined,
  CaretDownOutlined,
  TagOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import {
  setServerConfigMode,
  setVersionNumber,
  setIsLogin,
  setGithubBadge,
  setIsUnsafeMode,
} from '@/store/modules/globalSlice'
import showPng from '@/assets/show.png'
import logo from '@/assets/img/logo.svg'
import languageSvg from '@/assets/translate_language.svg'
import { Languages, LanguagesMap } from '@/utils/constants'
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
  const { isReadOnlyMode, versionNumber, isLogin, githubBadge, isUnsafeMode } =
    useSelector((state: any) => state.globalSlice)

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


  return (
    <div className={styles.wrapper}>
      {githubBadge && (
        <div>
          <a
            href="https://github.com/KusionStack/karpor"
            className={styles.github_corner}
            aria-label="View source on GitHub"
          >
            <svg
              width="80"
              height="80"
              viewBox="0 0 250 250"
              style={{
                fill: '#151513',
                color: '#fff',
                position: 'absolute',
                top: 0,
                border: 0,
                right: 0,
              }}
              aria-hidden="true"
            >
              <path d="M0,0 L115,115 L130,115 L142,142 L250,250 L250,0 Z"></path>
              <path
                d="M128.3,109.0 C113.8,99.7 119.0,89.6 119.0,89.6 C122.0,82.7 120.5,78.6 120.5,78.6 C119.2,72.0 123.4,76.3 123.4,76.3 C127.3,80.9 125.5,87.3 125.5,87.3 C122.9,97.6 130.6,101.9 134.4,103.2"
                fill="currentColor"
                style={{ transformOrigin: '130px 106px' }}
                className={styles.octo_arm}
              ></path>
              <path
                d="M115.0,115.0 C114.9,115.1 118.7,116.5 119.8,115.4 L133.7,101.6 C136.9,99.2 139.9,98.4 142.2,98.6 C133.8,88.0 127.5,74.4 143.8,58.0 C148.5,53.4 154.0,51.2 159.7,51.0 C160.3,49.4 163.2,43.6 171.4,40.1 C171.4,40.1 176.1,42.5 178.8,56.2 C183.1,58.6 187.2,61.8 190.9,65.4 C194.5,69.0 197.7,73.2 200.1,77.6 C213.8,80.2 216.3,84.9 216.3,84.9 C212.7,93.1 206.9,96.0 205.4,96.6 C205.1,102.4 203.0,107.8 198.3,112.5 C181.9,128.9 168.3,122.5 157.7,114.1 C157.9,116.9 156.7,120.9 152.7,124.9 L141.0,136.5 C139.8,137.7 141.6,141.9 141.8,141.8 Z"
                fill="currentColor"
                className={styles.octo_body}
              ></path>
            </svg>
          </a>
        </div>
      )}
      <div className={styles.nav}>
        <div className={styles.left}>
          <div className={styles.title} onClick={() => navigate('/')}>
            <div className={styles.sub_logo}>
              <img src={logo} alt="logo" />
            </div>
            <h4 className={styles.text}>Karpor</h4>
          </div>
          <Divider type="vertical" />
          <Menu
            style={{ flex: 1, border: 'none', fontSize: 13 }}
            mode="horizontal"
            selectedKeys={[pathname]}
            items={getMenuItems()}
            onClick={handleMenuClick}
          />
        </div>
        <div
          className={styles.right}
          style={githubBadge ? { marginRight: 80 } : {}}
        >
          <div className={styles.help}>
            <span
              onClick={() => {
                window.open('https://www.kusionstack.io/karpor')
              }}
            >
              <QuestionCircleOutlined style={{ color: '#646566' }} />
            </span>
          </div>
        </div>
      </div>
      <div className={styles.content}>
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
    </div>
  )
}

export default memo(LayoutPage)
