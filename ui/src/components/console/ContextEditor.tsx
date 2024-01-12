import { json, jsonParseLinter } from '@codemirror/lang-json';
import { linter } from '@codemirror/lint';
import { tokyoNight } from '@uiw/codemirror-theme-tokyo-night';
import CodeMirror from '@uiw/react-codemirror';
import React from 'react';

type ContextEditorProps = {
  id: string;
  setValue(value: string): void;
};

export const ContextEditor: React.FC<ContextEditorProps> = (
  props: ContextEditorProps
) => {
  const { setValue } = props;
  const onChange = React.useCallback(
    (val: any, _: any) => {
      setValue(val);
    },
    [setValue]
  );
  return (
    <CodeMirror
      value="{}"
      height="50vh"
      extensions={[json(), linter(jsonParseLinter())]}
      onChange={onChange}
      basicSetup={{
        lineNumbers: false,
        foldGutter: false,
        highlightActiveLineGutter: true,
        autocompletion: true,
        highlightActiveLine: true
      }}
      theme={tokyoNight}
    />
  );
};
