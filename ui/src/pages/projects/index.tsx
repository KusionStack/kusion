import React from 'react'
import CodeDiffView from '@/components/codeDiffView'
import { mockYaml, mockNewYaml } from '@/utils/tools'
import MarkdownView from '@/components/markdownView'

const Projects = () => {
  return (
      <>
        <CodeDiffView oldContent={mockYaml} newContent={mockNewYaml} />
      <div>
        <MarkdownView />
      </div>
      </>
  )
}

export default Projects
