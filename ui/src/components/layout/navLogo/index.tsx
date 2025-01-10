import React, { memo } from 'react'
import { useNavigate } from 'react-router-dom'
import logo from '@/assets/img/kusion_logo_white_transparent.png'

import styles from './style.module.less'

const NavLogo = () => {
  const navigate = useNavigate()

  return (
    <div className={styles.nav_logo} onClick={() => navigate('/')}>
      <div className={styles.nav_logo_img}>
        <img src={logo} alt="logo" />
      </div>
    </div>
  )
}

export default memo(NavLogo)
