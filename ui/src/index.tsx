import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import { Provider } from 'react-redux'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import enUS from 'antd/locale/en_US'
import { BrowserRouter } from 'react-router-dom'
import WrappedRoutes from '@/router'
import store from '@/store'

import '@/utils/request'

import './index.less'

dayjs.locale('zh-cn')

function App() {

  return (
    <Provider store={store}>
      <ConfigProvider
        locale={enUS}
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
