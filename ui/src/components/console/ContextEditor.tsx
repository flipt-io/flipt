import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import { useEffect, useRef, useState } from 'react';
import styles from './ContextEditor.module.css';
import './worker';

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

        return monaco.editor.create(monacoEl.current!, {
          lineNumbers: 'off',
          fontSize: 14,
          autoDetectHighContrast: true,
          folding: false,
          formatOnType: true,
          language: 'json',
          value: value || ['{', '\t"hello": "world"', '}'].join('\n')
        });
      });
    }
    // After initializing monaco editor
    if (editor) {
      subscription.current = editor.onDidChangeModelContent(() => {
        setValue(editor.getValue());
      });
    }
  }, [editor, value, setValue]);

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
