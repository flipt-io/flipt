import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';

// @ts-ignore
self.MonacoEnvironment = {
  getWorker(_, label) {
    switch (label) {
      case 'json':
        return new jsonWorker();
      default:
        return new editorWorker();
    }
  }
};
