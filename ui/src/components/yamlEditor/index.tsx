import React, { useEffect, useRef } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { StreamLanguage } from '@codemirror/language'
import { yaml } from '@codemirror/legacy-modes/mode/yaml'
import * as Themes from '@uiw/codemirror-themes-all'
import { highlightSelectionMatches, openSearchPanel } from '@codemirror/search';

type YamlEditorIProps = {
  value: string
  readOnly: boolean
  onChange?: (val: string) => void
  themeMode?: 'LIGHT' | 'DARK'
}

const YamlEditor = ({ value, readOnly = false, onChange, themeMode = 'DARK' }: YamlEditorIProps) => {

  const editorRef = useRef(null);

  useEffect(() => {
    if (editorRef.current?.view) {
      // const view = editorRef.current?.view;
      // openSearchPanel(view);
    }
  }, [editorRef.current?.view]);

  return (
    <div style={{ height: "100%", width: '100%' }}>
      <CodeMirror
        ref={editorRef}
        value={value}
        theme={themeMode === 'DARK' ? Themes.material : Themes.bbedit}
        height="100%"
        autoFocus
        onChange={onChange}
        extensions={[
          StreamLanguage.define(yaml),
          highlightSelectionMatches()
        ]}
        readOnly={readOnly}
        basicSetup={{
          lineNumbers: true,
          syntaxHighlighting: true,
        }}
      />
    </div>
  )
}

export default YamlEditor
