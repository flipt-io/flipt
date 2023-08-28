import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import tmrw from 'monaco-themes/themes/Tomorrow-Night-Bright.json';
import { useEffect, useRef } from 'react';
import styles from './ContextEditor.module.css';
import './worker';

type ContextEditorProps = {
  id: string;
  setValue(value: string): void;
};

export const ContextEditor: React.FC<ContextEditorProps> = (
  props: ContextEditorProps
) => {
  const { id, setValue } = props;

  const editor = useRef<undefined | monaco.editor.IStandaloneCodeEditor>();
  const monacoEl = useRef(null);
  const subscription = useRef<monaco.IDisposable | null>(null);

  monaco.editor.defineTheme('tmrw', tmrw as monaco.editor.IStandaloneThemeData);

  useEffect(() => {
    if (monacoEl.current) {
      editor.current = monaco.editor.create(monacoEl.current!, {
        value: '{}',
        language: 'json',
        fontSize: 14,
        lineNumbers: 'off',
        renderLineHighlight: 'none',
        minimap: {
          enabled: false
        },
        folding: false,
        autoDetectHighContrast: false,
        autoClosingBrackets: 'always',
        scrollBeyondLastLine: false,
        theme: 'tmrw'
      });

      // After initializing monaco editor
      if (editor?.current) {
        subscription.current = editor.current.onDidChangeModelContent(() => {
          if (editor?.current) setValue(editor.current.getValue());
        });
      }

      return () => editor.current && editor.current.dispose();
    }
  }, []);

  return <div id={id} className={styles.ContextEditor} ref={monacoEl}></div>;
};
