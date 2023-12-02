/* eslint-disable @typescript-eslint/no-use-before-define */
import { FlagType } from '~/types/Flag';

export const apiURL = '/api/v1';
export const authURL = '/auth/v1';
export const evaluateURL = '/evaluate/v1';
export const metaURL = '/meta';

const csrfTokenHeaderKey = 'x-csrf-token';
const sessionKey = 'session';

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
// TODO: find a better name for this
export function checkResponse(response: Response) {
  if (!response.ok) {
    if (response.status === 401) {
      window.localStorage.removeItem(csrfTokenHeaderKey);
      window.localStorage.removeItem(sessionKey);
      window.location.reload();
    }
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

export async function request(method: string, uri: string, body?: any) {
  const req = {
    method,
    headers: defaultHeaders(),
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

async function get(uri: string, base = apiURL) {
  return request('GET', base + uri);
}

async function post<T>(uri: string, values: T, base = apiURL) {
  return request('POST', base + uri, values);
}

async function put<T>(uri: string, values: T, base = apiURL) {
  return request('PUT', base + uri, values);
}

//
// auth
export async function listAuthMethods() {
  return get('/method', authURL);
}

export async function getAuthSelf() {
  return get('/self', authURL);
}

export async function expireAuthSelf() {
  return put('/self/expire', {}, authURL);
}

//
// evaluate
export async function evaluate(
  namespaceKey: string,
  flagKey: string,
  values: any
) {
  const body = {
    namespaceKey,
    flagKey,
    ...values
  };
  return post('/evaluate', body);
}

//
// evaluateV2
export async function evaluateV2(
  namespaceKey: string,
  flagKey: string,
  flagType: FlagType,
  values: any
) {
  const route = flagType === FlagType.BOOLEAN ? '/boolean' : '/variant';

  const body = {
    namespaceKey,
    flagKey: flagKey,
    ...values
  };
  return post(route, body, evaluateURL);
}

//
// meta
async function getMeta(path: string) {
  const req = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json'
    }
  };

  const res = await fetch(`${metaURL}${path}`, req);
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

export async function getInfo() {
  return getMeta('/info');
}

export async function getConfig() {
  return getMeta('/config');
}
