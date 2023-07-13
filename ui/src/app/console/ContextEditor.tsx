import Editor from '@monaco-editor/react';
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
    editorValue: editor.IModelContentChangedEvent
  ) => {
    field.onChange({ target: { value: value || '', id } });
  };

  return (
    <>
      <Editor
        height="30vh"
        defaultLanguage="json"
        language="json"
        defaultValue={['{', '\t"some": "context"', '}'].join('\n')}
        onChange={handleContextEditorChange}
      />
    </>
  );
}
