import { createApi } from '@reduxjs/toolkit/query/react';

import {
  IBatchFlagEvaluationCount,
  IFlagEvaluationCount
} from '~/types/Analytics';

import { internalQuery } from '~/utils/redux-rtk';

export const analyticsApi = createApi({
  reducerPath: 'analytics',
  baseQuery: internalQuery,
  tagTypes: ['Analytics'],
  endpoints: (builder) => ({
    // get evaluation count
    getFlagEvaluationCount: builder.query<
      IFlagEvaluationCount,
      {
        environmentKey: string;
        namespaceKey: string;
        flagKey: string;
        from?: string;
        to?: string;
      }
    >({
      query: ({ environmentKey, namespaceKey, flagKey, from, to }) => ({
        url: `/analytics/environments/${environmentKey}/namespaces/${namespaceKey}/flags/${flagKey}`,
        params: { from, to }
      })
    }),
    getBatchFlagEvaluationCount: builder.query<
      IBatchFlagEvaluationCount,
      {
        environmentKey: string;
        namespaceKey: string;
        flagKeys: string[];
        from?: string;
        to?: string;
        limit?: number;
      }
    >({
      query: ({ environmentKey, namespaceKey, flagKeys, from, to, limit }) => ({
        url: `/analytics/environments/${environmentKey}/namespaces/${namespaceKey}/batch`,
        method: 'POST',
        body: { flagKeys, from, to, limit }
      })
    })
  })
});

export const {
  useGetFlagEvaluationCountQuery,
  useGetBatchFlagEvaluationCountQuery
} = analyticsApi;
