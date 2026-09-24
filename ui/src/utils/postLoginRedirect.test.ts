/**
 * @jest-environment jsdom
 */
import {
  consumePostLoginRedirect,
  isSafeInAppRedirect,
  postLoginRedirectKey,
  postLoginRedirectMaxAgeMs,
  savePostLoginRedirect
} from './postLoginRedirect';

describe('post-login redirect round trip', () => {
  const store: Record<string, string> = {};
  const sessionStorageMock = {
    getItem: jest.fn((k: string) => store[k] ?? null),
    setItem: jest.fn((k: string, v: string) => {
      store[k] = v;
    }),
    removeItem: jest.fn((k: string) => {
      delete store[k];
    }),
    clear: jest.fn(() => {
      for (const k of Object.keys(store)) delete store[k];
    })
  };

  beforeEach(() => {
    sessionStorageMock.clear();
    sessionStorageMock.getItem.mockClear();
    sessionStorageMock.setItem.mockClear();
    sessionStorageMock.removeItem.mockClear();
    Object.defineProperty(window, 'sessionStorage', {
      value: sessionStorageMock,
      configurable: true
    });
  });

  it('round-trips a deep link through save and consume', () => {
    savePostLoginRedirect('/namespaces/default/flags/my-flag');
    expect(consumePostLoginRedirect()).toBe(
      '/namespaces/default/flags/my-flag'
    );
  });

  it('consumes only once', () => {
    savePostLoginRedirect('/flags');
    expect(consumePostLoginRedirect()).toBe('/flags');
    expect(consumePostLoginRedirect()).toBeNull();
  });

  it('normalizes /login to /', () => {
    savePostLoginRedirect('/login');
    expect(consumePostLoginRedirect()).toBe('/');
  });

  it('drops stale entries past the TTL', () => {
    const now = Date.now();
    window.sessionStorage.setItem(
      postLoginRedirectKey,
      JSON.stringify({
        target: '/flags',
        createdAt: now - postLoginRedirectMaxAgeMs - 1000
      })
    );
    expect(consumePostLoginRedirect(now)).toBeNull();
  });

  it('drops entries with a future timestamp', () => {
    const now = Date.now();
    window.sessionStorage.setItem(
      postLoginRedirectKey,
      JSON.stringify({ target: '/flags', createdAt: now + 60_000 })
    );
    expect(consumePostLoginRedirect(now)).toBeNull();
  });

  it('rejects absolute and protocol-relative URLs', () => {
    expect(isSafeInAppRedirect('https://evil.com/flags')).toBe(false);
    expect(isSafeInAppRedirect('//evil.com/flags')).toBe(false);
    expect(isSafeInAppRedirect('javascript:alert(1)')).toBe(false);
    expect(isSafeInAppRedirect('/back\\slash')).toBe(false);
    expect(isSafeInAppRedirect('/flags')).toBe(true);
  });

  it('drops a tampered absolute URL instead of redirecting to it', () => {
    window.sessionStorage.setItem(
      postLoginRedirectKey,
      JSON.stringify({ target: 'https://evil.com/', createdAt: Date.now() })
    );
    expect(consumePostLoginRedirect()).toBeNull();
  });

  it('tolerates a window.location-style hash-router value', () => {
    savePostLoginRedirect('/#/namespaces/default/flags/my-flag');
    expect(consumePostLoginRedirect()).toBe(
      '/namespaces/default/flags/my-flag'
    );
  });

  it('drops garbage and non-object payloads', () => {
    window.sessionStorage.setItem(postLoginRedirectKey, 'not-json{{{');
    expect(consumePostLoginRedirect()).toBeNull();
    window.sessionStorage.setItem(postLoginRedirectKey, '"just-a-string"');
    expect(consumePostLoginRedirect()).toBeNull();
  });
});
