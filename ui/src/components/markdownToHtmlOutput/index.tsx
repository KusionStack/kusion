import React from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { solarizedlight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import gfm from 'remark-gfm';

import { markdownString } from '@/utils/tools';

const MarkdownRenderer = ({ str = markdownString }) => {
  const markdown = `
# Hello, World!

This is a paragraph with **bold text** and *italic text*.

\`\`\`javascript
console.log("Hello, World!");
\`\`\`

## Here is a list:
- Item 1
- Item 2
- Item 3

[Link to React](https://reactjs.org)
`;

  return (
    <div>
      <h1>Markdown Output</h1>
      <ReactMarkdown
        children={str || markdown}
        remarkPlugins={[gfm]}
        components={{
          code({ node, className, children }) {
            const match = /language-(\w+)/.exec(className || '');
            return match ? (
              <SyntaxHighlighter style={solarizedlight} language={match[1]} PreTag="div">
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            ) : (
              <code className={className}>{children}</code>
            );
          }
        }}
      />
    </div>
  );
};

export default MarkdownRenderer;
