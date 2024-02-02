import {
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
  fetchBaseQuery
} from '@reduxjs/toolkit/query/react';
import { apiURL, checkResponse, defaultHeaders, internalURL } from '~/data/api';
type CustomFetchFn = (
  url: RequestInfo,
  options: RequestInit | undefined
) => Promise<Response>;

export const customFetchFn: CustomFetchFn = async (url, options) => {
  const headers = defaultHeaders();

  const response = await fetch(url, {
    ...options,
    headers
  });
  checkResponse(response);
  return response;
};

export const internalQuery = fetchBaseQuery({
  baseUrl: internalURL,
  fetchFn: customFetchFn
});

export const baseQuery: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  return fetchBaseQuery({
    baseUrl: apiURL,
    fetchFn: async (url, options) => {
      const state = api.getState();
      // @ts-ignore
      const ref = state?.refs?.currentRef;
      if (ref) {
        const req = url instanceof Request ? url : new Request(url);
        const q = new URLSearchParams({ reference: ref }).toString();
        const blob = req.headers.get('Content-Type')
          ? req.blob()
          : Promise.resolve(undefined);
        url = await blob.then(
          (body) =>
            new Request(req.url + '?' + q, {
              method: req.method,
              headers: req.headers,
              body: body,
              referrer: req.referrer,
              referrerPolicy: req.referrerPolicy,
              mode: req.mode,
              credentials: req.credentials,
              cache: req.cache,
              redirect: req.redirect,
              integrity: req.integrity
            })
        );
      }
      return customFetchFn(url, options);
    }
  })(args, api, extraOptions);
};
