import { fetchBaseQuery } from '@reduxjs/toolkit/query/react';
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

export const baseQuery = fetchBaseQuery({
  baseUrl: apiURL,
  fetchFn: customFetchFn
});

export const internalQuery = fetchBaseQuery({
  baseUrl: internalURL,
  fetchFn: customFetchFn
});
