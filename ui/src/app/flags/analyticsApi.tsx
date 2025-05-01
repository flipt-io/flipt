import { createApi } from '@reduxjs/toolkit/query/react';

import { IFlagEvaluationCount } from '~/types/Analytics';

import { internalV2Query } from '~/utils/redux-rtk';

export const analyticsApi = createApi({
  reducerPath: 'analytics',
  baseQuery: internalV2Query,
  tagTypes: ['Analytics'],
  endpoints: (builder) => ({
    // get evaluation count
    getFlagEvaluationCount: builder.query<
      IFlagEvaluationCount,
      { environmentKey: string; namespaceKey: string; flagKey: string; from: string; to: string }
    >({
      query: ({ environmentKey, namespaceKey, flagKey, from, to }) => ({
        url: `/analytics/environments/${environmentKey}/namespaces/${namespaceKey}/flags/${flagKey}`,
        params: { from, to }
      })
    })
  })
});

export const { useGetFlagEvaluationCountQuery } = analyticsApi;
