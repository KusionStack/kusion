import MarkdownRenderer from '@/components/markdownToHtmlOutput'
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
    <div>
      <Button onClick={() => setIsReadOnly(!isReadOnly)}>编辑</Button>
      <YamlEditor value={yamlValue} readOnly={isReadOnly} onChange={onChange} />
      <br />
      <MarkdownRenderer />
    </div>
  )
}

export default Insights
