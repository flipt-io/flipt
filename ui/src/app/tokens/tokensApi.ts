import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { authURL } from '~/data/api';
import {
  IAuthTokenBase,
  IAuthTokenInternalList,
  IAuthTokenSecret
} from '~/types/auth/Token';

import { customFetchFn } from '~/utils/redux-rtk';

export const tokensApi = createApi({
  reducerPath: 'tokens',
  baseQuery: fetchBaseQuery({
    baseUrl: authURL,
    fetchFn: customFetchFn
  }),
  tagTypes: ['Token'],
  endpoints: (builder) => ({
    // get list of tokens
    listTokens: builder.query<IAuthTokenInternalList, void>({
      query: () => {
        return { url: '/tokens', params: { method: 'METHOD_TOKEN' } };
      },
      providesTags: (_result, _error, _args) => [{ type: 'Token' }]
    }),
    // create a new token
    createToken: builder.mutation<IAuthTokenSecret, IAuthTokenBase>({
      query(values) {
        return {
          url: '/method/token',
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: () => [{ type: 'Token' }]
    }),
    // delete tokens
    deleteTokens: builder.mutation<void, string[]>({
      queryFn: async (ids, _api, _extraOptions, baseQuery) => {
        for (let id of ids) {
          const resp = await baseQuery({
            url: `/tokens/${id}`,
            method: 'DELETE'
          });
          if (resp.error) {
            return { error: resp.error };
          }
        }
        return { data: undefined };
      },
      invalidatesTags: () => [{ type: 'Token' }]
    })
  })
});

export const {
  useListTokensQuery,
  useCreateTokenMutation,
  useDeleteTokensMutation
} = tokensApi;
