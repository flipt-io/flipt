import * as helpers from './helpers';
import {
  buildAuthorizeUrl,
  redirectAfterLogout,
  resolvePostLoginRedirect
} from './navigation';

/**
 * @jest-environment jsdom
 */
jest.mock('./helpers', () => ({
  ...jest.requireActual('./helpers'),
  isSafeRedirectUrl: jest.fn()
}));

const mockRedirect = jest.fn();
const mockIsSafe = helpers.isSafeRedirectUrl as jest.Mock;

beforeEach(() => {
  jest.restoreAllMocks();
  mockRedirect.mockClear();
  mockIsSafe.mockReset();
});

it('calls redirect with the safe nextUri', () => {
  mockIsSafe.mockReturnValue(true);

  redirectAfterLogout(mockRedirect, {
    nextUri: 'https://auth-provider.com/logout'
  });

  expect(mockRedirect).toHaveBeenCalledWith(
    'https://auth-provider.com/logout',
    true
  );
});

it('calls redirect with the unsafe nextUri when isSafeRedirectUrl returns false', () => {
  mockIsSafe.mockReturnValue(false);

  redirectAfterLogout(
    mockRedirect,
    { nextUri: 'https://evil.com' },
    'my-issuer.com'
  );

  expect(mockRedirect).toHaveBeenCalledWith('//my-issuer.com', true);
});

it('calls redirect with /login when nextUri is missing and no issuer', () => {
  redirectAfterLogout(mockRedirect, {});

  expect(mockRedirect).toHaveBeenCalledWith('/login', false);
});

describe('buildAuthorizeUrl', () => {
  const origin = 'http://localhost:8080';

  it('returns the uri unchanged when no state is given', () => {
    expect(
      buildAuthorizeUrl('/auth/v1/method/github/authorize', undefined, origin)
    ).toBe('http://localhost:8080/auth/v1/method/github/authorize');
  });

  it('attaches the caller location as the state parameter', () => {
    expect(
      buildAuthorizeUrl('/auth/v1/method/github/authorize', '/flags', origin)
    ).toBe(
      'http://localhost:8080/auth/v1/method/github/authorize?state=%2Fflags'
    );
  });

  it('preserves existing query parameters', () => {
    expect(
      buildAuthorizeUrl(
        '/auth/v1/method/oidc/google/authorize?provider=google',
        '/namespaces/default/flags',
        origin
      )
    ).toBe(
      'http://localhost:8080/auth/v1/method/oidc/google/authorize?provider=google&state=%2Fnamespaces%2Fdefault%2Fflags'
    );
  });

  it('resolves an absolute uri against itself', () => {
    expect(
      buildAuthorizeUrl(
        'http://localhost:8080/auth/v1/method/github/authorize',
        '/analytics',
        origin
      )
    ).toBe(
      'http://localhost:8080/auth/v1/method/github/authorize?state=%2Fanalytics'
    );
  });
});

describe('resolvePostLoginRedirect', () => {
  it('returns the saved deep link including namespaced flag paths', () => {
    expect(resolvePostLoginRedirect('/namespaces/venus/flags/Flag1')).toBe(
      '/namespaces/venus/flags/Flag1'
    );
  });

  it('falls back to / when state is missing', () => {
    expect(resolvePostLoginRedirect(undefined)).toBe('/');
    expect(resolvePostLoginRedirect(null)).toBe('/');
    expect(resolvePostLoginRedirect('')).toBe('/');
  });

  it('falls back to / when state is not a string', () => {
    expect(resolvePostLoginRedirect({ pathname: '/flags' })).toBe('/');
  });

  it('rejects absolute and protocol-relative urls', () => {
    expect(resolvePostLoginRedirect('https://evil.example.com/flags')).toBe(
      '/'
    );
    expect(resolvePostLoginRedirect('//evil.example.com/flags')).toBe('/');
    expect(resolvePostLoginRedirect('/flags\\evil')).toBe('/');
  });
});
