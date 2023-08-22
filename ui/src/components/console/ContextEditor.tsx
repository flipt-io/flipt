import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import tmrw from 'monaco-themes/themes/Tomorrow-Night-Bright.json';
import { useEffect, useRef, useState } from 'react';
import styles from './ContextEditor.module.css';

type ContextEditorProps = {
  id: string;
  value: string | undefined;
  setValue(value: string): void;
};

export const ContextEditor: React.FC<ContextEditorProps> = (
  props: ContextEditorProps
) => {
  const { id, value, setValue } = props;

  const [editor, setEditor] =
    useState<monaco.editor.IStandaloneCodeEditor | null>(null);
  const monacoEl = useRef(null);

  const subscription = useRef<monaco.IDisposable | null>(null);

  useEffect(() => {
    if (monacoEl) {
      setEditor((editor) => {
        if (editor) return editor;

        monaco.editor.defineTheme(
          'tmrw',
          tmrw as monaco.editor.IStandaloneThemeData
        );

        return monaco.editor.create(monacoEl.current!, {
          value: ['{', '}'].join('\n'),
          language: 'json',
          fontSize: 14,
          lineNumbers: 'off',
          colorDecorators: true,
          renderLineHighlight: 'none',
          minimap: {
            enabled: false
          },
          contextmenu: false,
          folding: false,
          autoDetectHighContrast: false,
          autoClosingBrackets: 'always',
          scrollBeyondLastLine: false,
          theme: 'tmrw'
        });
      });
    }
    // After initializing monaco editor
    if (editor) {
      subscription.current = editor.onDidChangeModelContent(() => {
        setValue(editor.getValue());
      });
    }
  }, [editor, value, setValue, monacoEl]);

  useEffect(
    () => () => {
      if (editor) {
        editor?.dispose();
      }
      if (subscription.current) {
        subscription.current.dispose();
      }
    },
    []
  );

  return <div id={id} className={styles.ContextEditor} ref={monacoEl}></div>;
};
