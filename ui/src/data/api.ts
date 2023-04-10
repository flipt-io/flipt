import { IAuthTokenBase } from 'types/auth/Token';
import { IConstraintBase } from 'types/Constraint';
import { IDistributionBase } from 'types/Distribution';
import { IFlagBase } from 'types/Flag';
import { IRuleBase } from 'types/Rule';
import { ISegmentBase } from 'types/Segment';
import { IVariantBase } from 'types/Variant';

const apiURL = '/api/v1';
const authURL = '/auth/v1';
const metaURL = '/meta';
const csrfTokenHeaderKey = 'x-csrf-token';

export class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

//
// base methods
function setCsrf(req: any) {
  const csrfToken = window.localStorage.getItem(csrfTokenHeaderKey);
  if (csrfToken !== null) {
    req.headers[csrfTokenHeaderKey] = csrfToken;
  }

  return req;
}

export async function request(method: string, uri: string, body?: any) {
  const req = setCsrf({
    method,
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json'
    },
    body: JSON.stringify(body)
  });

  const res = await fetch(uri, req);
  if (!res.ok) {
    if (res.status === 401) {
      window.localStorage.clear();
      window.location.reload();
    }

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

export async function createToken(values: IAuthTokenBase) {
  return post('/method/token', values, authURL);
}

export async function deleteToken(id: string) {
  return del(`/tokens/${id}`, authURL);
}

export async function listTokens(method = 'METHOD_TOKEN') {
  return get(`/tokens?method=${method}`, authURL);
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

//
// rules
export async function listRules(namespaceKey: string, flagKey: string) {
  return get(`/namespaces/${namespaceKey}/flags/${flagKey}/rules`);
}

export async function createRule(
  namespaceKey: string,
  flagKey: string,
  values: IRuleBase
) {
  return post(`/namespaces/${namespaceKey}/flags/${flagKey}/rules`, values);
}

export async function deleteRule(
  namespaceKey: string,
  flagKey: string,
  ruleId: string
) {
  return del(`/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}`);
}

export async function orderRules(
  namespaceKey: string,
  flagKey: string,
  ruleIds: string[]
) {
  const req = setCsrf({
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ruleIds
    })
  });

  const res = await fetch(
    `${apiURL}/namespaces/${namespaceKey}/flags/${flagKey}/rules/order`,
    req
  );
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message);
  }
  return res.ok;
}

export async function createDistribution(
  namespaceKey: string,
  flagKey: string,
  ruleId: string,
  values: IDistributionBase
) {
  return post(
    `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}/distributions`,
    values
  );
}

export async function updateDistribution(
  namespaceKey: string,
  flagKey: string,
  ruleId: string,
  distributionId: string,
  values: IDistributionBase
) {
  return put(
    `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}/distributions/${distributionId}`,
    values
  );
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
// segments
export async function listSegments(namespaceKey: string) {
  return get(`/namespaces/${namespaceKey}/segments`);
}

export async function getSegment(namespaceKey: string, key: string) {
  return get(`/namespaces/${namespaceKey}/segments/${key}`);
}

export async function createSegment(
  namespaceKey: string,
  values: ISegmentBase
) {
  return post(`/namespaces/${namespaceKey}/segments`, values);
}

export async function updateSegment(
  namespaceKey: string,
  key: string,
  values: ISegmentBase
) {
  return put(`/namespaces/${namespaceKey}/segments/${key}`, values);
}

export async function deleteSegment(namespaceKey: string, key: string) {
  return del(`/namespaces/${namespaceKey}/segments/${key}`);
}

//
// constraints
export async function createConstraint(
  namespaceKey: string,
  segmentKey: string,
  values: IConstraintBase
) {
  return post(
    `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints`,
    values
  );
}

export async function updateConstraint(
  namespaceKey: string,
  segmentKey: string,
  constraintId: string,
  values: IConstraintBase
) {
  return put(
    `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`,
    values
  );
}

export async function deleteConstraint(
  namespaceKey: string,
  segmentKey: string,
  constraintId: string
) {
  return del(
    `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`
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
// meta
export async function getInfo() {
  const req = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json'
    }
  };

  const res = await fetch(`${metaURL}/info`, req);
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
