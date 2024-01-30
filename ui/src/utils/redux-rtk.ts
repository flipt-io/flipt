import {
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
  fetchBaseQuery
} from '@reduxjs/toolkit/query/react';
import { apiURL, checkResponse, defaultHeaders } from '~/data/api';
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
        // this is a hacky. It will only work for GET requests. Any other request will
        // be converted into get request and probably should be fine for ReadOnly storage.
        url = new Request(req.url + '?' + q);
      }
      return customFetchFn(url, options);
    }
  })(args, api, extraOptions);
};
