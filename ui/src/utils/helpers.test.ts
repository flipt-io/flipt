import {
  addNamespaceToPath,
  canFetchUpdates,
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

  it('should remove unsupported characters from generated keys', () => {
    const result = stringAsKey('My / Test #1!');
    expect(result).toEqual('my-test-1');
  });

  it('should preserve in-progress separator characters', () => {
    expect(stringAsKey('foo_')).toEqual('foo_');
    expect(stringAsKey('foo-bar')).toEqual('foo-bar');
    expect(stringAsKey('foo_-')).toEqual('foo_-');
  });
});

describe('canFetchUpdates', () => {
  it('should return true when user is authenticated', () => {
    const session = { authenticated: true, required: true };
    expect(canFetchUpdates(session)).toBe(true);
  });

  it('should return false when user is not authenticated and auth is required', () => {
    const session = { authenticated: false, required: true };
    expect(canFetchUpdates(session)).toBe(false);
  });

  it('should return true when auth is not required', () => {
    const session = { authenticated: false, required: false };
    expect(canFetchUpdates(session)).toBe(true);
  });

  it('should return false when session is undefined', () => {
    expect(canFetchUpdates(undefined)).toBe(false);
  });

  it('should return false when session is null', () => {
    expect(canFetchUpdates(null)).toBe(false);
  });
});
