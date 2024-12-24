import React from 'react'
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import remarkGfm from 'remark-gfm';
import { dark, vscDarkPlus, coyWithoutShadows, darcula } from 'react-syntax-highlighter/dist/esm/styles/prism'

import { copy, markdownString } from "@/utils/tools"

import styles from "./styles.module.less"
import { Button } from 'antd';
import { data } from '@remix-run/router';
import { CopyOutlined } from '@ant-design/icons';

type MarkdownViewIProps = {
  markdown?: string
  themeMode?: 'LIGHT' | 'DARK'
}

const them = {
  DARK: vscDarkPlus,
  LIGHT: coyWithoutShadows
};

const MarkdownView = ({ markdown = markdownString, themeMode = 'DARK' }: MarkdownViewIProps) => {
  console.log("=====MarkdownView=====")
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
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
    </div>
  );

}

export default MarkdownView
