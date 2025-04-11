/* eslint-disable @typescript-eslint/no-use-before-define */
import { FlagType } from '~/types/Flag';
import { IAuthGithubInternal } from '~/types/auth/Github';
import { IAuthJWTInternal } from '~/types/auth/JWT';
import { IAuthOIDCInternal } from '~/types/auth/OIDC';

import { getUser } from '~/data/user';

export const apiURL = 'v2/environments';
export const authURL = 'auth/v1';
export const evaluateURL = 'evaluate/v1';
export const internalURL = 'internal/v1';
export const metaURL = 'meta/info';

export const sessionKey = 'session';

const csrfTokenHeaderKey = 'x-csrf-token';
const fliptEnvironmentHeaderKey = 'x-flipt-environment';

export type Session = {
  required: boolean;
  authenticated: boolean;
  self?: IAuthOIDCInternal | IAuthGithubInternal | IAuthJWTInternal;
};

export class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

//
// base methods
function headerCsrf(): Record<string, string> {
  const csrfToken = window.localStorage.getItem(csrfTokenHeaderKey);
  if (csrfToken !== null) {
    return { 'x-csrf-token': csrfToken };
  }
  return {};
}

export function checkResponse(response: Response) {
  if (!response.ok && response.status === 401) {
    const session = window.localStorage.getItem(sessionKey);
    window.localStorage.removeItem(csrfTokenHeaderKey);
    window.localStorage.removeItem(sessionKey);
    if (session) {
      try {
        const user = getUser(JSON.parse(session));
        if (user?.issuer) {
          window.location.href = `//${user.issuer}`;
          return;
        }
      } catch (e) {
        //
      }
    }
    window.location.reload();
  }
}

export function defaultHeaders(): Record<string, string> {
  const headers = {
    ...headerCsrf(),
    'Content-Type': 'application/json',
    Accept: 'application/json',
    'Cache-Control': 'no-store'
  };

  return headers;
}

export async function request(
  method: string,
  uri: string,
  body?: any,
  headers?: Record<string, string>
  ) {
  const req = {
    method,
    headers: { ...defaultHeaders(), ...headers },
    body: JSON.stringify(body)
  };

  const res = await fetch(uri, req);

  if (!res.ok) {
    checkResponse(res);
    const contentType = res.headers.get('content-type');

    if (!contentType || !contentType.includes('application/json')) {
      const err = new APIError('An unexpected error occurred.', res.status);
      throw err;
    }

    let err = await res.json();
    throw new APIError(err.message, res.status);
  }
  return res.json();
}

async function get(uri: string, base = apiURL, headers?: Record<string, string>) {
  return request('GET', base + uri, undefined, headers);
}

async function post<T>(uri: string, values: T, base = apiURL, headers?: Record<string, string>) {
  return request('POST', base + uri, values, headers);
}

async function put<T>(uri: string, values: T, base = apiURL, headers?: Record<string, string>) {
  return request('PUT', base + uri, values, headers);
}

//
// auth
export async function getAuthSelf() {
  return get('/self', authURL);
}

export async function expireAuthSelf() {
  return put('/self/expire', {}, authURL);
}

//
// evaluate
export async function evaluate(
  environmentName: string,
  namespaceKey: string,
  flagKey: string,
  flagType: FlagType,
  values: any
) {
  let route = flagType === FlagType.BOOLEAN ? '/boolean' : '/variant';
  const body = {
    namespaceKey,
    flagKey: flagKey,
    ...values
  };
  return post(route, body, evaluateURL, {
    [fliptEnvironmentHeaderKey]: environmentName
  });
}

//
// meta
export async function getMeta() {
  const req = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json'
    }
  };

  const res = await fetch(metaURL, req);
  if (!res.ok) {
    const contentType = res.headers.get('content-type');

    if (!contentType || !contentType.includes('application/json')) {
      const err = new APIError('An unexpected error occurred.', res.status);
      throw err;
    }

    let err = await res.json();
    throw new APIError(err.message, res.status);
  }

  const token = res.headers.get(csrfTokenHeaderKey);
  if (token !== null) {
    window.localStorage.setItem(csrfTokenHeaderKey, token);
  }

  return res.json();
}
