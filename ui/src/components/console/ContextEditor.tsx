import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import 'monaco-editor/esm/vs/language/json/monaco.contribution';
import tmrw from 'monaco-themes/themes/Tomorrow-Night-Bright.json';
import { useEffect, useRef } from 'react';
import styles from './ContextEditor.module.css';
import './worker';

monaco.editor.defineTheme('tmrw', tmrw as monaco.editor.IStandaloneThemeData);

type ContextEditorProps = {
  id: string;
  setValue(value: string): void;
};

export const ContextEditor: React.FC<ContextEditorProps> = (
  props: ContextEditorProps
) => {
  const { id, setValue } = props;

  const monacoEl = useRef(null);

  useEffect(() => {
    if (monacoEl.current) {
      const editor = monaco.editor.create(monacoEl.current!, {
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
      const subscription = editor.onDidChangeModelContent(() => {
        setValue(editor.getValue());
      });

      return () => {
        subscription.dispose();
        editor.dispose();
      };
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return <div id={id} className={styles.ContextEditor} ref={monacoEl}></div>;
};
