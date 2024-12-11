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

import { createClient, createConfig } from '@hey-api/client-fetch';
import { client as kusionClient, SourceService } from '@kusionstack/kusion-api-client-sdk';

const client = createClient({
  baseUrl: 'http://30.177.51.253:80',
});

console.log(client, "====client======")

// client.setConfig({
//   baseUrl: 'http://30.177.51.253:80'
// });

dayjs.locale('zh-cn')

function App() {

  async function example() {
    try {
      const sources = await SourceService.listSource();
      console.log('Sources:', sources.data);
    } catch (error) {
      console.error('Error:', error);
    }
  }

  useEffect(() => {
    example();
  }, [])

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
