import React from 'react'
import PageContainer from '@/components/pageContainer'
import CodeDiffView from '@/components/codeDiffView'
import { mockYaml, mockNewYaml } from '@/utils/tools'
import MarkdownView from '@/components/markdownView'

const Projects = () => {
  return (
    <PageContainer title="Projects">
      <CodeDiffView oldContent={mockYaml} newContent={mockNewYaml} />
      <div>
        <MarkdownView />
      </div>
    </PageContainer>
  )
}

export default Projects
