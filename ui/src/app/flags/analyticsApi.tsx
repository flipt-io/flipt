import { createApi } from '@reduxjs/toolkit/query/react';
import { IFlagEvaluationCount } from '~/types/Analytics';
import { internalQuery } from '~/utils/redux-rtk';

export const analyticsApi = createApi({
  reducerPath: 'analytics',
  baseQuery: internalQuery,
  tagTypes: ['Analytics'],
  endpoints: (builder) => ({
    // get evaluation count
    getFlagEvaluationCount: builder.query<
      IFlagEvaluationCount,
      { namespaceKey: string; flagKey: string; from: string; to: string }
    >({
      query: ({ namespaceKey, flagKey, from, to }) => ({
        url: `/analytics/namespaces/${namespaceKey}/flags/${flagKey}`,
        params: { from, to }
      })
    })
  })
});

export const { useGetFlagEvaluationCountQuery } = analyticsApi;
