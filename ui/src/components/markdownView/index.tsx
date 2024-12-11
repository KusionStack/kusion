import React from 'react'
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
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
    <ReactMarkdown
      components={{
        code({ node, className, children, ...props }) {
          console.log(node, className, props, "=====sadsadasd===")
          const match = /language-(\w+)/.exec(className || '');
          return match ? (
            <SyntaxHighlighter
              showLineNumbers={true}
              style={them?.[themeMode]}
              language={match[1]}
              PreTag='div'
              {...props}
            >
              {String(children).replace(/\n$/, '')}
            </SyntaxHighlighter>
          ) : (
            <code className={className} {...props}>
              {children}
            </code>
          );
        }
      }}
    >
      {markdown}
    </ReactMarkdown>
  );

}

export default MarkdownView
