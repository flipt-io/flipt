import { addNamespaceToPath, titleCase, upperFirst } from './helpers';

describe('addNamespaceToPath', () => {
  it('should return flags path for root path', () => {
    const path = '/';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example/flags');
  });

  it('should return flags path for same namespace', () => {
    const path = '/namespaces/example';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example/flags');
  });

  it('should preserve flags section when switching namespace', () => {
    const path = '/namespaces/example/flags';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/flags');
  });

  it('should preserve flags section with trailing slash when switching namespace', () => {
    const path = '/namespaces/example/flags/';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/flags');
  });

  it('should drop item key and preserve flags section when switching namespace', () => {
    const path = '/namespaces/example/flags/my-flag';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/flags');
  });

  it('should preserve segments section when switching namespace', () => {
    const path = '/namespaces/example/segments';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/segments');
  });

  it('should drop item key and preserve segments section when switching namespace', () => {
    const path = '/namespaces/example/segments/my-segment';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/segments');
  });

  it('should preserve console section when switching namespace', () => {
    const path = '/namespaces/example/console';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/console');
  });

  it('should default to flags for non-namespaced paths', () => {
    const path = '/settings';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example/flags');
  });

  it('should default to flags for unknown namespace sections', () => {
    const path = '/namespaces/example/unknown';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/flags');
  });

  it('should drop nested subroutes and preserve section', () => {
    const path = '/namespaces/example/flags/my-flag/rules';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/flags');
  });
});

describe('upperFirst', () => {
  it('should convert first char to upper case', () => {
    const result = upperFirst('test is done');
    expect(result).toEqual('Test is done');
  });
});

describe('titleCase', () => {
  it('should convert first char to upper case for each word', () => {
    const result = titleCase('test is done');
    expect(result).toEqual('Test Is Done');
  });
});
