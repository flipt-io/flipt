import { Diagnostic } from '@codemirror/lint';
import { EditorView } from '@codemirror/view';
import { Text } from '@codemirror/state';
import { SearchCursor } from '@codemirror/search';

function getJsonErrorPosition(error: SyntaxError, doc: Text): number {
  let m;
  if ((m = error.message.match(/at position (\d+)/)))
    return Math.min(+m[1], doc.length);
  if ((m = error.message.match(/at line (\d+) column (\d+)/)))
    return Math.min(doc.line(+m[1]).from + +m[2] - 1, doc.length);
  return 0;
}

export const parseLinter =
  () =>
  (view: EditorView): Diagnostic[] => {
    const doc = view.state.doc;
    try {
      const data = JSON.parse(doc.toString());
      if (Array.isArray(data) || typeof data !== 'object') {
        return [
          {
            from: 0,
            message: 'Request Context: unexpected array',
            severity: 'error',
            to: doc.length
          }
        ];
      }
      for (const [key, value] of Object.entries(data)) {
        if (typeof value !== 'string') {
          const keyCursor = new SearchCursor(doc, `"${key}"`);
          const keyPosition = keyCursor.next().value;
          const valueCursor = new SearchCursor(
            doc,
            JSON.stringify(value),
            keyPosition.to
          );
          const pos = valueCursor.next().value;
          return [
            {
              from: pos.from,
              message: 'Request Context: expected string value',
              severity: 'error',
              to: pos.to
            }
          ];
        }
      }
    } catch (e) {
      if (!(e instanceof SyntaxError)) throw e;
      const pos = getJsonErrorPosition(e, doc);
      return [
        {
          from: pos,
          message: 'JSON: unexpected keyword',
          severity: 'error',
          to: pos
        }
      ];
    }
    return [];
  };
