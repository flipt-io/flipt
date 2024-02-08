import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { authURL } from '~/data/api';
import { IAuthMethodList } from '~/types/Auth';

export const authProvidersApi = createApi({
  reducerPath: 'providers',
  baseQuery: fetchBaseQuery({
    baseUrl: authURL
  }),
  tagTypes: ['Provider'],
  endpoints: (builder) => ({
    // get list of tokens
    listAuthProviders: builder.query<IAuthMethodList, void>({
      query: () => {
        return { url: '/method' };
      },
      providesTags: (_result, _error, _args) => [{ type: 'Provider' }]
    })
  })
});

export const { useListAuthProvidersQuery } = authProvidersApi;
