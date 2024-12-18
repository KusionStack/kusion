import React, { memo, useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import {
  setServerConfigMode,
  setVersionNumber,
  setGithubBadge,
  setIsUnsafeMode,
} from '@/store/modules/globalSlice'
import { useAxios } from '@/utils/request'
import NavLogo from './navLogo'
import NavMenu from './navMenu'
import NavRight from './navRight'

import styles from './style.module.less'


const LayoutPage = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const dispatch = useDispatch()
  const { isLogin, isUnsafeMode } = useSelector((state: any) => state.globalSlice)

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
      <div className={styles.nav}>
        <div className={styles.nav_left}>
          <NavLogo />
          <NavMenu />
        </div>
        <NavRight />
      </div>
      <div className={styles.body_wrapper}>
        <Outlet />
      </div>
    </div>
  )
}

export default memo(LayoutPage)
