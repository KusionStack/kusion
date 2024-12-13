import React from 'react'
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import remarkGfm from 'remark-gfm';
import { dark, vscDarkPlus, coyWithoutShadows, darcula } from 'react-syntax-highlighter/dist/esm/styles/prism'

import { markdownString } from "@/utils/tools"

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
    <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
  );

}

export default MarkdownView
