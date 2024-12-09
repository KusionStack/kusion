import React from 'react'
// import { AutoComplete, Input, message, Space, Tag } from 'antd'
// import {
//   DoubleLeftOutlined,
//   DoubleRightOutlined,
//   CloseOutlined,
// } from '@ant-design/icons'

// import styles from './styles.module.less'
import { Tabs } from 'antd'
import SourcesPanelContent from './sourcePanelContent'
import ModulePanelContent from './modulePanelContent'
import PageContainer from '@/components/pageContainer'

const Modules = () => {
  // const [searchType, setSearchType] = useState<string>('sql')

  // function handleTabChange(value: string) {
  //   setSearchType(value)
  // }

  const tabsItmes = [
    {
      label: 'Sources',
      key: 'Sources',
      children: <SourcesPanelContent />,
    },
    {
      label: 'Modules',
      key: 'Modules',
      children: <ModulePanelContent />,
    },
  ]

  return (
    <PageContainer title="Modules">
      <Tabs defaultActiveKey="1" type="card" items={tabsItmes as any} />
    </PageContainer>
  )
}

export default Modules
