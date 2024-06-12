/**
 * @jest-environment jsdom
 * @jest-environment-options {"url": "https://test/"}
 */

import { checkResponse, sessionKey } from './api';

describe('checkResponse', () => {
  let originalLocation: Location;
  beforeEach(() => {
    originalLocation = window.location;
    delete (window as any).location;
    (window as any).location = {
      href: '',
      reload: () => {}
    };

    Object.defineProperty(window.location, 'href', {
      writable: true,
      value: 'https://test/'
    });
    jest.spyOn(window.location, 'reload').mockImplementation(() => {});
  });

  afterEach(() => {
    window.location = originalLocation;
    jest.restoreAllMocks();
  });

  it('should not change the state', () => {
    window.localStorage.setItem(sessionKey, 'value');
    checkResponse({ status: 200, ok: true } as Response);
    expect(window.location.href).toEqual('https://test/');
    checkResponse({ status: 404, ok: false } as Response);
    expect(window.location.href).toEqual('https://test/');
    expect(window.localStorage.getItem(sessionKey)).toEqual('value');
  });

  it('should reload if method oauth', () => {
    window.localStorage.setItem(
      sessionKey,
      '{"authenticated":true,"required":true,"self":{"method":"METHOD_GITHUB", "metadata": {}}}'
    );
    checkResponse({ status: 401, ok: false } as Response);
    expect(window.location.href).toEqual('https://test/');
    expect(window.localStorage.getItem(sessionKey)).toBe(null);
    expect(window.location.reload).toHaveBeenCalled();
  });

  it('should redirect back to issuer for jwt auth method', () => {
    window.localStorage.setItem(
      sessionKey,
      '{"authenticated":true,"required":true,"self":{"method":"METHOD_JWT", "metadata": {"io.flipt.auth.jwt.issuer":"flipt.issuer"}}}'
    );
    checkResponse({ status: 401, ok: false } as Response);
    expect(window.location.href).toEqual('//flipt.issuer');
    expect(window.localStorage.getItem(sessionKey)).toBe(null);
  });
});
