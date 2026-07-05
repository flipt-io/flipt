/**
 * @jest-environment jsdom
 * @jest-environment-options {"url": "https://test/"}
 */
import * as api from './api';

function mockFetch<T>(status: number, body: T): jest.Mock {
  return jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: new Map([['content-type', 'application/json']]),
    json: async () => body
  });
}

describe('expireAuthSelf', () => {
  beforeEach(() => {
    jest.spyOn(api.browser, 'reloadPage').mockImplementation(() => {});
    jest.spyOn(api.browser, 'navigateTo').mockImplementation(() => {});
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('should call DELETE /auth/v1/self/revoke', async () => {
    const mock = mockFetch(200, { nextUri: 'https://provider/logout' });
    global.fetch = mock;

    const result = await api.revokeAuthSelf();

    expect(mock).toHaveBeenCalledWith(
      'auth/v1/self/revoke',
      expect.objectContaining({
        method: 'DELETE'
      })
    );
    expect(result).toEqual({ nextUri: 'https://provider/logout' });
  });

  it('should succeed without nextUri', async () => {
    global.fetch = mockFetch(200, {});

    const result = await api.revokeAuthSelf();

    expect(result).toEqual({});
  });
});

describe('checkResponse', () => {
  const defaultURL = 'https://test/';

  beforeEach(() => {
    window.history.replaceState({}, '', defaultURL);
    jest.spyOn(api.browser, 'reloadPage').mockImplementation(() => {});
    jest.spyOn(api.browser, 'navigateTo').mockImplementation(() => {});
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('should not change the state', () => {
    window.localStorage.setItem(api.sessionKey, 'value');

    api.checkResponse({ status: 200, ok: true } as Response);
    api.checkResponse({ status: 404, ok: false } as Response);

    expect(window.location.href).toEqual(defaultURL);
    expect(window.localStorage.getItem(api.sessionKey)).toEqual('value');
    expect(api.browser.reloadPage).not.toHaveBeenCalled();
    expect(api.browser.navigateTo).not.toHaveBeenCalled();
  });

  it('should reload if method oauth', () => {
    window.localStorage.setItem(
      api.sessionKey,
      '{"authenticated":true,"required":true,"self":{"method":"METHOD_GITHUB", "metadata": {}}}'
    );

    api.checkResponse({ status: 401, ok: false } as Response);

    expect(window.location.href).toEqual(defaultURL);
    expect(window.localStorage.getItem(api.sessionKey)).toBe(null);
    expect(api.browser.reloadPage).toHaveBeenCalled();
    expect(api.browser.navigateTo).not.toHaveBeenCalled();
  });

  it('should redirect back to issuer for jwt auth method', () => {
    window.localStorage.setItem(
      api.sessionKey,
      '{"authenticated":true,"required":true,"self":{"method":"METHOD_JWT", "metadata": {"io.flipt.auth.jwt.issuer":"flipt.issuer"}}}'
    );

    api.checkResponse({ status: 401, ok: false } as Response);

    expect(api.browser.navigateTo).toHaveBeenCalledWith('//flipt.issuer');
    expect(api.browser.reloadPage).not.toHaveBeenCalled();
    expect(window.localStorage.getItem(api.sessionKey)).toBe(null);
  });
});
