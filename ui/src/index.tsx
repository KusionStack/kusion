import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import { Provider } from 'react-redux'
import dayjs from 'dayjs'
import enUS from 'antd/locale/en_US'
import { BrowserRouter } from 'react-router-dom'
import { client } from '@kusionstack/kusion-api-client-sdk';
import WrappedRoutes from '@/router'
import store from '@/store'

import './index.less'


client.setConfig({
  baseUrl: 'http://30.177.35.25'
});

dayjs.locale('en-US')

function App() {

  return (
    <Provider store={store}> 
      <ConfigProvider
        locale={enUS}
        theme={{
          token: {
            colorPrimary: '#2667FF',
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
