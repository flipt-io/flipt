import * as helpers from './helpers';
import { redirectAfterLogout } from './navigation';

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
