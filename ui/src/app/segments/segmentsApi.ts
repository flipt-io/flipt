import { TagDescription } from '@reduxjs/toolkit/query';
import { createApi } from '@reduxjs/toolkit/query/react';
import { SortingState } from '@tanstack/react-table';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '~/store';
import { IConstraintBase } from '~/types/Constraint';
import { ISegment, ISegmentBase, ISegmentList } from '~/types/Segment';
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

export const segmentTag = (namespaceKey: string): TagDescription<'Segment'> => {
  return { type: 'Segment', id: namespaceKey };
};

export const segmentsApi = createApi({
  reducerPath: 'segments',
  baseQuery,
  tagTypes: ['Segment'],
  endpoints: (builder) => ({
    // get list of segments in this namespace
    listSegments: builder.query<ISegmentList, string>({
      query: (namespaceKey) => `/namespaces/${namespaceKey}/segments`,
      providesTags: (result, _error, namespaceKey) =>
        result
          ? [
              ...result.segments.map(({ key }) => ({
                type: 'Segment' as const,
                id: namespaceKey + '/' + key
              })),
              segmentTag(namespaceKey)
            ]
          : [segmentTag(namespaceKey)]
    }),
    // get segment in this namespace
    getSegment: builder.query<
      ISegment,
      { namespaceKey: string; segmentKey: string }
    >({
      query: ({ namespaceKey, segmentKey }) =>
        `/namespaces/${namespaceKey}/segments/${segmentKey}`,
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
        segmentTag(namespaceKey),
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
          url: `/namespaces/${namespaceKey}/segments/${segmentKey}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey }) => [
        segmentTag(namespaceKey)
      ]
    }),
    // update the segment in the namespace
    updateSegment: builder.mutation<
      ISegment,
      { namespaceKey: string; segmentKey: string; values: ISegmentBase }
    >({
      query({ namespaceKey, segmentKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/segments/${segmentKey}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, segmentKey }) => [
        segmentTag(namespaceKey),
        { type: 'Segment', id: namespaceKey + '/' + segmentKey }
      ]
    }),
    // copy the segment from one namespace to another one
    copySegment: builder.mutation<
      void,
      {
        from: { namespaceKey: string; segmentKey: string };
        to: { namespaceKey: string; segmentKey: string };
      }
    >({
      queryFn: async ({ from, to }, _api, _extraOptions, baseQuery) => {
        let resp = await baseQuery({
          url: `/namespaces/${from.namespaceKey}/segments/${from.segmentKey}`,
          method: 'get'
        });
        if (resp.error) {
          return { error: resp.error };
        }
        let data = resp.data as ISegment;

        if (to.segmentKey) {
          data.key = to.segmentKey;
        }
        // first create the segment
        resp = await baseQuery({
          url: `/namespaces/${to.namespaceKey}/segments`,
          method: 'POST',
          body: data
        });
        if (resp.error) {
          return { error: resp.error };
        }
        // then copy the constraints
        const constraints = data.constraints || [];
        for (let constraint of constraints) {
          resp = await baseQuery({
            url: `/namespaces/${to.namespaceKey}/segments/${to.segmentKey}/constraints`,
            method: 'POST',
            body: constraint
          });
          if (resp.error) {
            return { error: resp.error };
          }
        }

        return { data: undefined };
      },
      invalidatesTags: (_result, _error, { to }) => [
        segmentTag(to.namespaceKey)
      ]
    }),

    // create the segment constraint in the namespace
    createConstraint: builder.mutation<
      void,
      { namespaceKey: string; segmentKey: string; values: IConstraintBase }
    >({
      query({ namespaceKey, segmentKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints`,
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
          url: `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`,
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
          url: `/namespaces/${namespaceKey}/segments/${segmentKey}/constraints/${constraintId}`,
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
  useCopySegmentMutation,
  useCreateConstraintMutation,
  useUpdateConstraintMutation,
  useDeleteConstraintMutation
} = segmentsApi;
