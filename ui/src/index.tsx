import React, { useEffect } from 'react'
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

import { client } from '@kusionstack/kusion-api-client-sdk';



client.setConfig({
  baseUrl: 'http://30.177.52.72:80'
});

dayjs.locale('zh-cn')

console.log(client, "====client======")

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
