import React from 'react'
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Button } from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import { copy } from "@/utils/tools"

import styles from "./styles.module.less"

type MarkdownViewIProps = {
  markdown?: string
  themeMode?: 'LIGHT' | 'DARK'
}

const MarkdownView = ({ markdown }: MarkdownViewIProps) => {
  return (
    <div className={styles.markdown_view_container}>
      <div className={styles.markdown_view_copy}>
        <Button
          type="primary"
          size="small"
          onClick={() => copy(markdown)}
          disabled={!markdown}
          icon={<CopyOutlined />}
        >
          copy
        </Button>
      </div>
      <ReactMarkdown className={styles.markdown_container} remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
    </div>
  );

}

export default MarkdownView
