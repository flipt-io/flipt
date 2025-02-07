import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { SortingState } from '@tanstack/react-table';
import { v4 as uuid } from 'uuid';

import { IConstraint } from '~/types/Constraint';
import { IResourceListResponse, IResourceResponse } from '~/types/Resource';
import { ISegment, ISegmentList } from '~/types/Segment';

import { RootState } from '~/store';
import { baseQuery } from '~/utils/redux-rtk';

const initialTableState: {
  sorting: SortingState;
} = {
  sorting: []
};

export const segmentsTableSlice = createSlice({
  name: 'segmentsTable',
  initialState: initialTableState,
  reducers: {
    setSorting: (state, action: PayloadAction<SortingState>) => {
      const newSorting = action.payload;
      state.sorting = newSorting;
    }
  }
});

export const selectSorting = (state: RootState) => state.segmentsTable.sorting;
export const { setSorting } = segmentsTableSlice.actions;

export const segmentsApi = createApi({
  reducerPath: 'segments',
  baseQuery,
  tagTypes: ['Segment'],
  endpoints: (builder) => ({
    // get list of segments in this namespace
    listSegments: builder.query<
      ISegmentList,
      { environmentKey: string; namespaceKey: string }
    >({
      query: ({ environmentKey, namespaceKey }) =>
        `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Segment`,
      providesTags: (result, _error, { environmentKey, namespaceKey }) =>
        result
          ? [
              ...result.segments.map(({ key }) => ({
                type: 'Segment' as const,
                id: environmentKey + '/' + namespaceKey + '/' + key
              })),
              { type: 'Segment', id: environmentKey + '/' + namespaceKey }
            ]
          : [{ type: 'Segment', id: environmentKey + '/' + namespaceKey }],
      transformResponse: (
        response: IResourceListResponse<ISegment>
      ): ISegmentList => {
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
        return {
          segments: response.resources.map(({ payload }) => payload)
        } as ISegmentList;
      }
    }),
    // get segment in this namespace
    getSegment: builder.query<
      ISegment,
      { environmentKey: string; namespaceKey: string; segmentKey: string }
    >({
      query: ({ environmentKey, namespaceKey, segmentKey }) =>
        `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Segment/${segmentKey}`,
      providesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, segmentKey }
      ) => [
        {
          type: 'Segment',
          id: environmentKey + '/' + namespaceKey + '/' + segmentKey
        }
      ],
      transformResponse: (response: IResourceResponse<ISegment>): ISegment => {
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
        return {
          ...response.resource.payload,
          constraints: response.resource.payload.constraints?.map(
            (c: IConstraint) => ({
              ...c,
              id: uuid()
            })
          )
        };
      }
    }),
    // create a new segment in the namespace
    createSegment: builder.mutation<
      ISegment,
      {
        environmentKey: string;
        namespaceKey: string;
        values: ISegment;
        revision?: string;
      }
    >({
      query({ environmentKey, namespaceKey, values, revision }) {
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources`,
          method: 'POST',
          body: {
            key: values.key,
            revision: revision,
            payload: {
              '@type': 'flipt.core.Segment',
              ...values
            }
          }
        };
      },
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, values }
      ) => [
        { type: 'Segment', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Segment',
          id: environmentKey + '/' + namespaceKey + '/' + values.key
        }
      ]
    }),
    // delete the segment from the namespace
    deleteSegment: builder.mutation<
      void,
      {
        environmentKey: string;
        namespaceKey: string;
        segmentKey: string;
        revision: string;
      }
    >({
      query({ environmentKey, namespaceKey, segmentKey, revision }) {
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Segment/${segmentKey}?revision=${revision}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, segmentKey }
      ) => [
        { type: 'Segment', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Segment',
          id: environmentKey + '/' + namespaceKey + '/' + segmentKey
        }
      ]
    }),
    // update the segment in the namespace
    updateSegment: builder.mutation<
      ISegment,
      {
        environmentKey: string;
        namespaceKey: string;
        segmentKey: string;
        values: ISegment;
        revision: string;
      }
    >({
      query({ environmentKey, namespaceKey, segmentKey, values, revision }) {
        const payload = {
          '@type': 'flipt.core.Segment',
          ...values
        };
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources`,
          method: 'PUT',
          body: {
            key: segmentKey,
            revision,
            payload
          }
        };
      },
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, segmentKey }
      ) => [
        { type: 'Segment', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Segment',
          id: environmentKey + '/' + namespaceKey + '/' + segmentKey
        }
      ]
    }),
    // copy the segment from one namespace to another one
    copySegment: builder.mutation<
      void,
      {
        environmentKey: string;
        from: { namespaceKey: string; segmentKey: string };
        to: { namespaceKey: string; segmentKey: string };
      }
    >({
      queryFn: async (
        { environmentKey, from, to },
        _api,
        _extraOptions,
        baseQuery
      ) => {
        let resp = await baseQuery({
          url: `/${environmentKey}/namespaces/${from.namespaceKey}/resources/flipt.core.Segment/${from.segmentKey}`,
          method: 'GET'
        });
        if (resp.error) {
          return { error: resp.error };
        }

        const res = resp.data as {
          resource: { payload: ISegment; key: string };
          revision: string;
        };

        let data = {
          key: res.resource.key,
          payload: res.resource.payload,
          revision: res.revision
        };

        resp = await baseQuery({
          url: `/${environmentKey}/namespaces/${to.namespaceKey}/resources`,
          method: 'POST',
          body: data
        });
        if (resp.error) {
          return { error: resp.error };
        }
        return { data: undefined };
      },
      invalidatesTags: (_result, _error, { environmentKey, to }) => [
        { type: 'Segment', id: environmentKey + '/' + to.namespaceKey }
      ]
    })
  })
});

export const {
  useListSegmentsQuery,
  useGetSegmentQuery,
  useCreateSegmentMutation,
  useDeleteSegmentMutation,
  useUpdateSegmentMutation,
  useCopySegmentMutation
} = segmentsApi;
