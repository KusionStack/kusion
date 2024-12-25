import React, { memo, useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import NavLogo from './navLogo'
import KusionMenu from './kusionMenu'
import NavRight from './navRight'

import styles from './style.module.less'


const LayoutPage = () => {

  return (
    <div className={styles.wrapper}>
      <div className={styles.nav}>
        <div className={styles.nav_left}>
          <NavLogo />
          {/* <NavMenu /> */}
          <KusionMenu />
          
        </div>
        <NavRight />
      </div>
      <div className={styles.body_wrapper}>
        {/* <Card>
          
        </Card> */}
        <Outlet />
      </div>
    </div>
  )
}

export default memo(LayoutPage)
