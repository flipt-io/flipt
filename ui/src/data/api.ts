/* eslint-disable @typescript-eslint/no-use-before-define */
import { FlagType, IFlagBase } from '~/types/Flag';
import { IVariantBase } from '~/types/Variant';

export const apiURL = '/api/v1';
const authURL = '/auth/v1';
const evaluateURL = '/evaluate/v1';
const metaURL = '/meta';
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

async function del(uri: string, base = apiURL) {
  return request('DELETE', base + uri);
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
// namespaces
export async function listNamespaces() {
  return get('/namespaces');
}

export async function getNamespace(key: string) {
  return get(`/namespaces/${key}`);
}

export async function createNamespace(values: any) {
  return post('/namespaces', values);
}

export async function updateNamespace(key: string, values: any) {
  return put(`/namespaces/${key}`, values);
}

export async function deleteNamespace(key: string) {
  return del(`/namespaces/${key}`);
}

//
// flags
export async function listFlags(namespaceKey: string) {
  return get(`/namespaces/${namespaceKey}/flags`);
}

export async function getFlag(namespaceKey: string, key: string) {
  return get(`/namespaces/${namespaceKey}/flags/${key}`);
}

export async function createFlag(namespaceKey: string, values: IFlagBase) {
  return post(`/namespaces/${namespaceKey}/flags`, values);
}

export async function updateFlag(
  namespaceKey: string,
  key: string,
  values: IFlagBase
) {
  return put(`/namespaces/${namespaceKey}/flags/${key}`, values);
}

export async function deleteFlag(namespaceKey: string, key: string) {
  return del(`/namespaces/${namespaceKey}/flags/${key}`);
}

export async function copyFlag(
  from: { namespaceKey: string; key: string },
  to: { namespaceKey: string; key?: string }
) {
  let flag = await get(`/namespaces/${from.namespaceKey}/flags/${from.key}`);
  if (to.key) {
    flag.key = to.key;
  }

  // first create the flag
  await post(`/namespaces/${to.namespaceKey}/flags`, flag);

  // then copy the variants
  for (let variant of flag.variants) {
    await createVariant(to.namespaceKey, flag.key, variant);
  }
}
//
// variants
export async function createVariant(
  namespaceKey: string,
  flagKey: string,
  values: IVariantBase
) {
  return post(`/namespaces/${namespaceKey}/flags/${flagKey}/variants`, values);
}

export async function updateVariant(
  namespaceKey: string,
  flagKey: string,
  variantId: string,
  values: IVariantBase
) {
  return put(
    `/namespaces/${namespaceKey}/flags/${flagKey}/variants/${variantId}`,
    values
  );
}

export async function deleteVariant(
  namespaceKey: string,
  flagKey: string,
  variantId: string
) {
  return del(
    `/namespaces/${namespaceKey}/flags/${flagKey}/variants/${variantId}`
  );
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
