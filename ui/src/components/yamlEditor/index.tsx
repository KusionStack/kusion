import React from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { StreamLanguage } from '@codemirror/language'
import { yaml } from '@codemirror/legacy-modes/mode/yaml'
import * as Themes from '@uiw/codemirror-themes-all'

type YamlEditorIProps = {
  value: string
  readOnly: boolean
  onChange?: (val: string) => void
  themeMode?: 'LIGHT' | 'DARK'
}

const YamlEditor = ({ value = '', readOnly = false, onChange, themeMode = 'DARK' }: YamlEditorIProps) => {

  return (
    <div style={{ height: "100%", width: '100%' }}>
      <CodeMirror
        value={value}
        theme={themeMode === 'DARK' ? Themes.material : Themes.bbedit}
        height="100%"
        onChange={onChange}
        extensions={[StreamLanguage.define(yaml)]}
        readOnly={readOnly}
      />
    </div>
  )
}

export default YamlEditor
