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

const portAPI = 'api/server-port';
const defaultPort = '80';

async function loadServerConfig() {
  try {
    const response = await fetch(portAPI);
    const config = await response.json();
    const port = config?.port || defaultPort;

    client.setConfig({
      baseUrl: `http://localhost:${port}`
    });
  } catch (error) {
    client.setConfig({
      baseUrl: `http://localhost:${defaultPort}`
    });
  }
}

loadServerConfig();

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
