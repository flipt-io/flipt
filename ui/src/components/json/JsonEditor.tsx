import { json } from '@codemirror/lang-json';
import { linter, lintGutter } from '@codemirror/lint';
import { tokyoNight } from '@uiw/codemirror-theme-tokyo-night';
import CodeMirror from '@uiw/react-codemirror';
import React from 'react';
import { parseLinter } from './lint';

type JsonEditorProps = {
  id: string;
  value: string;
  setValue(value: string): void;
  disabled?: boolean;
};

export const JsonEditor: React.FC<JsonEditorProps> = (
  props: JsonEditorProps
) => {
  const { value = '{}', setValue, disabled = false } = props;
  const onChange = React.useCallback(
    (val: any, _: any) => {
      setValue(val);
    },
    [setValue]
  );
  return (
    <CodeMirror
      value={value}
      height="50vh"
      extensions={[json(), lintGutter(), linter(parseLinter())]}
      onChange={onChange}
      readOnly={disabled}
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
