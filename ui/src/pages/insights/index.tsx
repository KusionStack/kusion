import React, { useState } from 'react'
import { Button } from 'antd'
import YamlEditor from '@/components/yamlEditor'
import CodeDiffView from '@/components/codeDiffView'
import MarkdownView from '@/components/markdownView'
import CodeMirrorMarkdown from '@/components/codeMirrorMarkdown'
import { mockYaml, mockNewYaml } from '@/utils/tools'
import G6Tree from '@/components/g6Tree'

const Insights = () => {
  const [yamlValue, setYamlValue] = useState(mockYaml)
  const [isReadOnly, setIsReadOnly] = useState(true)

  function onChange(value) {
    setYamlValue(value)
  }
  return (
    <div>
      <G6Tree />
      <Button onClick={() => setIsReadOnly(!isReadOnly)}>编辑</Button>
      <YamlEditor value={yamlValue} readOnly={isReadOnly} onChange={onChange} />
      <br />
      <>
        <CodeDiffView oldContent={mockYaml} newContent={mockNewYaml} />
        <div>
          <MarkdownView />
          <br />
          <CodeMirrorMarkdown />
        </div>
      </>
    </div>
  )
}

export default Insights
