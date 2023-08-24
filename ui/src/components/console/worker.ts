// @ts-ignore
self.MonacoEnvironment = {
  getWorker(_moduleId, _label) {
    const worker = new Worker(
      new URL(
        'monaco-editor/esm/vs/language/json/json.worker.js',
        import.meta.url
      )
    );
    return worker;
  }
};
