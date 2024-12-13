import React from 'react'
import CodeDiffView from '@/components/codeDiffView'
import { mockYaml, mockNewYaml } from '@/utils/tools'
import MarkdownView from '@/components/markdownView'
import CodeMirrorMarkdown from '@/components/codeMirrorMarkdown'

const Projects = () => {
  return (
      <>
        <CodeDiffView oldContent={mockYaml} newContent={mockNewYaml} />
      <div>
        <MarkdownView />
        <br/>
        <CodeMirrorMarkdown/>
      </div>
      </>
  )
}

export default Projects
