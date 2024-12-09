import React from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { StreamLanguage } from '@codemirror/language'
import { yaml } from '@codemirror/legacy-modes/mode/yaml'
import * as Themes from '@uiw/codemirror-themes-all'

const YamlEditor = ({ value = '', readOnly = false, onChange }) => {

  return (
    <div style={{ height: "100%", width: '100%' }}>
      <CodeMirror
        value={value}
        theme={Themes.materialDark}
        height="100%"
        onChange={onChange}
        extensions={[StreamLanguage.define(yaml)]}
        readOnly={readOnly}
      />
    </div>
  )
}

export default YamlEditor
