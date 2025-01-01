import { json } from '@codemirror/lang-json';
import { linter } from '@codemirror/lint';
import { tokyoNight } from '@uiw/codemirror-theme-tokyo-night';
import CodeMirror from '@uiw/react-codemirror';
import React from 'react';
import { parseLinter } from './lint';

type JsonEditorProps = {
  id: string;
  value: string;
  setValue(value: string): void;
  disabled?: boolean;
  strict?: boolean;
  height?: string;
  'data-testid'?: string;
};

export const JsonEditor: React.FC<JsonEditorProps> = (
  props: JsonEditorProps
) => {
  const {
    value = '{}',
    setValue,
    disabled = false,
    strict = true,
    height = '50vh',
    'data-testid': dataTestId
  } = props;
  const onChange = React.useCallback(
    (val: any, _: any) => {
      setValue(val);
    },
    [setValue]
  );
  return (
    <CodeMirror
      value={value}
      height={height}
      extensions={[json(), linter(parseLinter(strict))]}
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
      data-testid={dataTestId}
    />
  );
};
