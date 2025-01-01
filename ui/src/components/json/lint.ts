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
  (strict: boolean = true) =>
  (view: EditorView): Diagnostic[] => {
    const doc = view.state.doc;
    try {
      const data = JSON.parse(doc.toString());
      if (strict && Array.isArray(data)) {
        return [
          {
            from: 0,
            message: 'JSON: unexpected array',
            severity: 'error',
            to: doc.length
          }
        ];
      }
      if (!Array.isArray(data) && typeof data !== 'object') {
        return [
          {
            from: 0,
            message: 'JSON: unexpected type',
            severity: 'error',
            to: doc.length
          }
        ];
      }
      if (!Array.isArray(data)) {
        for (const [key, value] of Object.entries(data)) {
          if (strict && typeof value !== 'string') {
            const keyCursor = new SearchCursor(doc, `"${key}"`);
            const keyMatch = keyCursor.next();
            if (!keyMatch?.value) continue;

            const valueCursor = new SearchCursor(
              doc,
              JSON.stringify(value),
              keyMatch.value.to
            );
            const valueMatch = valueCursor.next();
            if (!valueMatch?.value) continue;

            return [
              {
                from: valueMatch.value.from,
                message: 'JSON: expected string value',
                severity: 'error',
                to: valueMatch.value.to
              }
            ];
          }
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
