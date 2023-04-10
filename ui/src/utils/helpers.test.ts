import { addNamespaceToPath } from './helpers';

describe('addNamespaceToPath', () => {
  it('should return a path with the namespace key', () => {
    const path = '/';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example/');
  });

  it('should leave a path with the same namespace key alone', () => {
    const path = '/namespaces/example';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example');
  });

  it('handles existing noun in path', () => {
    const path = '/segments';
    const key = 'example';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/example/segments');
  });

  it('handles existing noun in path with new key', () => {
    const path = '/namespaces/example/segments';
    const key = 'test';
    const result = addNamespaceToPath(path, key);
    expect(result).toEqual('/namespaces/test/segments');
  });
});
