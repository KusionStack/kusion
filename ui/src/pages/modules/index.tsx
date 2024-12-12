import React, { useEffect } from 'react'
import { Tabs } from 'antd'
import SourcesPanelContent from './sourcePanelContent'
import ModulePanelContent from './modulePanelContent'

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
      <Tabs defaultActiveKey="1" type="card" items={tabsItmes as any} />
  )
}

export default Modules
