import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { IAuthMethodList } from '~/types/Auth';

export const authProvidersApi = createApi({
  reducerPath: 'providers',
  baseQuery: fetchBaseQuery({
    baseUrl: '/auth/v1'
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
