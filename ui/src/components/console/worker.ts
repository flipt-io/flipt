// @ts-ignore
self.MonacoEnvironment = {
  getWorker(_, label) {
    switch (label) {
      case 'json':
        return new Worker(
          new URL(
            'monaco-editor/esm/vs/language/json/json.worker.js',
            import.meta.url
          )
        );
      default:
        return new Worker(
          new URL(
            'monaco-editor/esm/vs/editor/editor.worker.js',
            import.meta.url
          )
        );
    }
  }
};
