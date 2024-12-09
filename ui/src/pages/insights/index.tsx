import PageContainer from '@/components/pageContainer'
import YamlEditor from '@/components/yamlEditor'
import { mockYaml } from '@/utils/tools'
import { Button } from 'antd'
import React, { useState } from 'react'

const Insights = () => {
  const [yamlValue, setYamlValue] = useState(mockYaml)
  const [isReadOnly, setIsReadOnly] = useState(true)

  function onChange(value) {
    setYamlValue(value)
  }
  return (
    <PageContainer title="Insights">
      <div>
        <Button onClick={() => setIsReadOnly(!isReadOnly)}>编辑</Button>
        <YamlEditor value={yamlValue} readOnly={isReadOnly} onChange={onChange} />
      </div>
    </PageContainer>
  )
}

export default Insights
