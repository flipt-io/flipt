import Editor, { Monaco } from '@monaco-editor/react';
import { useField } from 'formik';
import { editor } from 'monaco-editor';

type ContextEditorProps = {
  id: string;
  name: string;
};

export default function ContextEditor(props: ContextEditorProps) {
  const { id } = props;
  const [field] = useField(props);

  const handleContextEditorChange = (
    value: string | undefined,
    _editorValue: editor.IModelContentChangedEvent
  ) => {
    field.onChange({ target: { value: value || '', id } });
  };

  const handleEditorDidMount = (
    editor: editor.IStandaloneCodeEditor,
    monaco: Monaco
  ) => {
    import('monaco-themes/themes/Tomorrow-Night.json').then((data) => {
      monaco.editor.defineTheme(
        'tomorrow-night',
        data as editor.IStandaloneThemeData
      );
      monaco.editor.setTheme('tomorrow-night');
    });
  };

  return (
    <>
      <Editor
        height="30vh"
        defaultLanguage="json"
        language="json"
        value={['{', '\t"some": "context"', '}'].join('\n')}
        options={{
          lineNumbers: 'off',
          scrollbar: {
            vertical: 'hidden'
          },
          minimap: {
            enabled: false
          },
          tabSize: 2
        }}
        onChange={handleContextEditorChange}
        onMount={handleEditorDidMount}
      />
    </>
  );
}
