import React from 'react'
import ReactDiffViewer, { DiffMethod } from 'react-diff-viewer-continued'
import YamlEditor from '../yamlEditor'
import styles from './style.module.less'

const CodeDiffView = ({ oldContent, newContent }) => {

  const diffStyles = {
    variables: {
      dark: {
        highlightBackground: '#fefed5',
        highlightGutterBackground: '#ffcd3c',
      },
    },
    line: {
      padding: '10px 2px',
      '&:hover': {
        background: '#a26ea1',
      },
    },
  }

  return (
    <div className={styles.kusion_code_diff}>
      {newContent ? (
        <div className={styles.kusion_code_diff_content}>
          <ReactDiffViewer
            leftTitle={'Before'}
            rightTitle={'After'}
            styles={diffStyles}
            oldValue={oldContent}
            newValue={newContent}
            splitView={true}
            useDarkTheme={false}
            compareMethod={DiffMethod.LINES}
          />
        </div>
      ) : (
        <div className={styles.kusion_code_diff_before}>
          <div className={styles.kusion_code_diff_before_title}>{'Before'}</div>
          <YamlEditor readOnly={true} value={oldContent} themeMode={'LIGHT'}/>
        </div>
      )}
    </div>
  )
}

export default CodeDiffView
