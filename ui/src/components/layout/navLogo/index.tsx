import React, { memo } from 'react'
import { useNavigate } from 'react-router-dom'
import logo from '@/assets/img/logo.svg'

import styles from './style.module.less'

const NavLogo = () => {
  const navigate = useNavigate()

  return (
    <div className={styles.nav_logo} onClick={() => navigate('/')}>
      <div className={styles.nav_logo_img}>
        <img src={logo} alt="logo" />
      </div>
      <h4 className={styles.nav_logo_title}>Karpor</h4>
    </div>
  )
}

export default memo(NavLogo)
