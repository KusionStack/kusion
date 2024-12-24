import React, { memo, useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import NavLogo from './navLogo'
import NavMenu from './navMenu'
import NavRight from './navRight'

import styles from './style.module.less'


const LayoutPage = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const { isLogin, isUnsafeMode } = useSelector((state: any) => state.globalSlice)


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
