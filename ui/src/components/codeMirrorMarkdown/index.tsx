import React from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { EditorView } from '@codemirror/view';
import { markdown } from '@codemirror/lang-markdown';
import * as Themes from '@uiw/codemirror-themes-all'

import { copy, markdownString } from "@/utils/tools"
import { Button } from 'antd';
import { CopyOutlined } from '@ant-design/icons';

import styles from "./styles.module.less"

// 假设你已经有了一个 Markdown 字符串
// const markdownContent = `
// # Hello, world!
// This is a **Markdown** example.
// `;

function CodeMirrorMarkdown() {
  return (
    <div className={styles.codemirror_markdown_container}>
      <div className={styles.codemirror_markdown_copy}>
        {markdownString && (
          <Button
            type="primary"
            size="small"
            onClick={() => copy(markdownString)}
            disabled={!markdownString}
            icon={<CopyOutlined />}
          >
            Copy
          </Button>
        )}
      </div>
      <CodeMirror
        value={markdownString}
        extensions={[
          markdown(),
          EditorView.lineWrapping,
        ]}
        theme={Themes?.darcula}
        readOnly // 设置为只读模式以便于展示
      />
    </div>
  );
}

export default CodeMirrorMarkdown;

