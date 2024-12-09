import React, { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import { Provider } from 'react-redux'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import { useTranslation } from 'react-i18next'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import deDE from 'antd/locale/de_DE'
import ptBR from 'antd/locale/pt_BR'
import { BrowserRouter } from 'react-router-dom'
import WrappedRoutes from '@/router'
import store from '@/store'
import './i18n'

import '@/utils/request'

import './index.less'

dayjs.locale('zh-cn')

function App() {
  const { i18n } = useTranslation()
  const currentLocale = localStorage.getItem('lang')
  const [lang, setLang] = useState(currentLocale || 'en')

  useEffect(() => {
    setLang(i18n.language)
  }, [i18n.language])

  const langMap = {
    en: enUS,
    zh: zhCN,
    de: deDE,
    pt: ptBR,
  }

  return (
    <Provider store={store}>
      <ConfigProvider
        locale={langMap?.[lang || 'en']}
        theme={{
          token: {
            colorPrimary: '#2f54eb',
          },
        }}
      >
        <BrowserRouter>
          <WrappedRoutes />
        </BrowserRouter>
      </ConfigProvider>
    </Provider>
  )
}

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)
root.render(<App />)
