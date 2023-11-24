import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { IConstraintBase } from '~/types/Constraint';
import { ISegment, ISegmentBase, ISegmentList } from '~/types/Segment';

export const segmentsApi = createApi({
  reducerPath: 'segments',
  // TODO: there are no headers for production
  baseQuery: fetchBaseQuery({ baseUrl: '/api/v1/' }),
  tagTypes: ['Segment'],
  endpoints: (builder) => ({
    // get list of segments in this namespace
    listSegments: builder.query<ISegmentList, string>({
      query: (namespaceKey) => `namespaces/${namespaceKey}/segments`,
      providesTags: (result, _error, namespaceKey) =>
        result
          ? [
              ...result.segments.map(({ key }) => ({
                type: 'Segment' as const,
                id: namespaceKey + '/' + key
              })),
              { type: 'Segment', id: namespaceKey }
            ]
          : [{ type: 'Segment', id: namespaceKey }]
    }),
    // get segment in this namespace
    getSegment: builder.query<
      ISegment,
      { namespaceKey: string; segmentKey: string }
    >({
      query: ({ namespaceKey, segmentKey }) =>
        `namespaces/${namespaceKey}/segments/${segmentKey}`,
      providesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // create a new segment in the namespace
    createSegment: builder.mutation<
      ISegment,
      { namespaceKey: string; values: ISegmentBase }
    >({
      query({ namespaceKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/segments`,
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, values }) => [
        { type: 'Segment', id: namespaceKey },
        { type: 'Segment', id: namespaceKey + '/' + values.key }
      ]
    }),
    // delete the segment from the namespace
    deleteSegment: builder.mutation<
      void,
      { namespaceKey: string; segmentKey: string }
    >({
      query({ namespaceKey, segmentKey }) {
        return {
          url: `namespaces/${namespaceKey}/segments/${segmentKey}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey },
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // update the segment in the namespace
    updateSegment: builder.mutation<
      ISegment,
      { namespaceKey: string; segmentKey: string; values: ISegmentBase }
    >({
      query({ namespaceKey, segmentKey, values }) {
        return {
          url: `namespaces/${namespaceKey}/segments/${segmentKey}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey },
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // create the segment constraint in the namespace
    createConstraint: builder.mutation<
      void,
      { namespaceKey: string; segmentKey: string; values: IConstraintBase }
    >({
      query({ namespaceKey, segmentKey, values }) {
        return {
          url: `namespaces/${namespaceKey}/segments/${segmentKey}/constraints`,
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // update the segment constraint in the namespace
    updateConstraint: builder.mutation<
      void,
      {
        namespaceKey: string;
        segmentKey: string;
        constraintId: string;
        values: IConstraintBase;
      }
    >({
      query({ namespaceKey, segmentKey, constraintId, values }) {
        return {
          url: `namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // delete the segment constraint in the namespace
    deleteConstraint: builder.mutation<
      void,
      { namespaceKey: string; segmentKey: string; constraintId: string }
    >({
      query({ namespaceKey, segmentKey, constraintId }) {
        return {
          url: `namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
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
  useCreateConstraintMutation,
  useUpdateConstraintMutation,
  useDeleteConstraintMutation
} = segmentsApi;
