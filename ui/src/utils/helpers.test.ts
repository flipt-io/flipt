import {
  addNamespaceToPath,
  stringAsKey,
  titleCase,
  upperFirst
} from './helpers';

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

describe('stringAsKey', () => {
  it('should convert a string to a key with spaces and lowercase', () => {
    const result = stringAsKey('test is done');
    expect(result).toEqual('test-is-done');
  });

  it('should convert a string to a key with spaces and uppercase', () => {
    const result = stringAsKey('Test Is Done');
    expect(result).toEqual('test-is-done');
  });
});
