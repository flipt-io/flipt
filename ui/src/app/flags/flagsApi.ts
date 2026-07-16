import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { SortingState } from '@tanstack/react-table';

import { revisionChanged } from '~/app/environments/environmentsApi';

import { IFlag, IFlagList } from '~/types/Flag';
import { INamespaceList } from '~/types/Namespace';
import { IResourceListResponse, IResourceResponse } from '~/types/Resource';
import { IRollout } from '~/types/Rollout';
import { IRule } from '~/types/Rule';

import { RootState } from '~/store';
import { baseQuery } from '~/utils/redux-rtk';
import { uuid } from '~/utils/uuid';

const initialTableState: {
  sorting: SortingState;
} = {
  sorting: []
};

export const flagsTableSlice = createSlice({
  name: 'flagsTable',
  initialState: initialTableState,
  reducers: {
    setSorting: (state, action: PayloadAction<SortingState>) => {
      const newSorting = action.payload;
      state.sorting = newSorting;
    }
  }
});

export const selectSorting = (state: RootState) => state.flagsTable.sorting;
export const { setSorting } = flagsTableSlice.actions;

function enrichFlag(flag: IFlag): IFlag {
  return {
    ...flag,
    rollouts: flag.rollouts?.map((r: IRollout, i: number) => ({
      ...r,
      id: uuid(),
      rank: i
    })),
    rules: flag.rules?.map((r: IRule, i: number) => ({
      ...r,
      id: uuid(),
      rank: i
    }))
  };
}

export const flagsApi = createApi({
  reducerPath: 'flags',
  baseQuery,
  tagTypes: ['Flag'],
  endpoints: (builder) => ({
    // get list of flags in this namespace
    listFlags: builder.query<
      IFlagList,
      { environmentKey: string; namespaceKey: string }
    >({
      query: ({ environmentKey, namespaceKey }) =>
        `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag`,
      providesTags: (result, _error, { environmentKey, namespaceKey }) =>
        result
          ? [
              ...result.flags.map(({ key }) => ({
                type: 'Flag' as const,
                id: environmentKey + '/' + namespaceKey + '/' + key
              })),
              { type: 'Flag', id: environmentKey + '/' + namespaceKey }
            ]
          : [{ type: 'Flag', id: environmentKey + '/' + namespaceKey }],
      transformResponse: (
        response: IResourceListResponse<IFlag>
      ): IFlagList => {
        return {
          flags: response.resources.map(({ payload }) => payload)
        } as IFlagList;
      }
    }),
    // get flag in this namespace
    getFlag: builder.query<
      IFlag,
      { environmentKey: string; namespaceKey: string; flagKey: string }
    >({
      query: ({ environmentKey, namespaceKey, flagKey }) =>
        `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag/${flagKey}`,
      providesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, flagKey }
      ) => [
        {
          type: 'Flag',
          id: environmentKey + '/' + namespaceKey + '/' + flagKey
        },
        { type: 'Flag', id: environmentKey + '/' + namespaceKey }
      ],
      transformResponse: (response: IResourceResponse<IFlag>): IFlag => {
        return enrichFlag(response.resource.payload);
      }
    }),
    // create a new flag in the namespace
    createFlag: builder.mutation<
      IResourceResponse<IFlag>,
      {
        environmentKey: string;
        namespaceKey: string;
        values: IFlag;
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
              '@type': 'flipt.core.Flag',
              ...values
            }
          }
        };
      },
      async onQueryStarted(
        { environmentKey, namespaceKey },
        { dispatch, queryFulfilled }
      ) {
        try {
          const { data, meta } = await queryFulfilled;
          if (meta?.revision) {
            dispatch(
              revisionChanged({ environmentKey, revision: meta.revision })
            );
          }
          const payload = data.resource.payload;
          dispatch(
            flagsApi.util.upsertQueryData(
              'getFlag',
              {
                environmentKey,
                namespaceKey,
                flagKey: data.resource.key
              },
              enrichFlag(payload)
            )
          );
          dispatch(
            flagsApi.util.updateQueryData(
              'listFlags',
              { environmentKey, namespaceKey },
              (draft) => {
                draft.flags.push(payload);
              }
            )
          );
        } catch {
          // Mutation failed, no cache updates needed
        }
      }
    }),
    // delete the flag from the namespace
    deleteFlag: builder.mutation<
      void,
      {
        environmentKey: string;
        namespaceKey: string;
        flagKey: string;
        revision: string;
      }
    >({
      query({ environmentKey, namespaceKey, flagKey, revision }) {
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag/${flagKey}?revision=${revision}`,
          method: 'DELETE'
        };
      },
      async onQueryStarted(
        { environmentKey, namespaceKey, flagKey },
        { dispatch, queryFulfilled }
      ) {
        try {
          const { meta } = await queryFulfilled;
          if (meta?.revision) {
            dispatch(
              revisionChanged({ environmentKey, revision: meta.revision })
            );
          }
          dispatch(
            flagsApi.util.updateQueryData(
              'listFlags',
              { environmentKey, namespaceKey },
              (draft) => {
                draft.flags = draft.flags.filter((f) => f.key !== flagKey);
              }
            )
          );
        } catch {
          // Mutation failed, no cache updates needed
        }
      }
    }),
    // update the flag in the namespace
    updateFlag: builder.mutation<
      IResourceResponse<IFlag>,
      {
        environmentKey: string;
        namespaceKey: string;
        flagKey: string;
        values: IFlag;
        revision?: string;
      }
    >({
      query({ environmentKey, namespaceKey, flagKey, values, revision }) {
        const payload = {
          '@type': 'flipt.core.Flag',
          ...values
        };
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources`,
          method: 'PUT',
          body: {
            key: flagKey,
            revision,
            payload
          }
        };
      },
      async onQueryStarted(
        { environmentKey, namespaceKey, flagKey },
        { dispatch, queryFulfilled }
      ) {
        try {
          const { data, meta } = await queryFulfilled;
          if (meta?.revision) {
            dispatch(
              revisionChanged({ environmentKey, revision: meta.revision })
            );
          }
          const payload = data.resource.payload;
          dispatch(
            flagsApi.util.upsertQueryData(
              'getFlag',
              { environmentKey, namespaceKey, flagKey },
              enrichFlag(payload)
            )
          );
          dispatch(
            flagsApi.util.updateQueryData(
              'listFlags',
              { environmentKey, namespaceKey },
              (draft) => {
                const idx = draft.flags.findIndex((f) => f.key === flagKey);
                if (idx >= 0) {
                  draft.flags[idx] = payload;
                }
              }
            )
          );
        } catch {
          // Mutation failed, no cache updates needed
        }
      }
    }),
    // copy the flag from one namespace to another one
    copyFlag: builder.mutation<
      void,
      {
        from: {
          environmentKey: string;
          namespaceKey: string;
          flagKey: string;
        };
        to: {
          environmentKey: string;
          namespaceKey: string;
          flagKey: string;
        };
      }
    >({
      queryFn: async ({ from, to }, _api, _extraOptions, baseQuery) => {
        let resp = await baseQuery({
          url: `/${from.environmentKey}/namespaces/${from.namespaceKey}/resources/flipt.core.Flag/${from.flagKey}`,
          method: 'GET'
        });
        if (resp.error) {
          return { error: resp.error };
        }

        const res = resp.data as {
          resource: { payload: IFlag; key: string };
          revision: string;
        };

        resp = await baseQuery({
          url: `/${to.environmentKey}/namespaces`,
          method: 'GET'
        });
        if (resp.error) {
          return { error: resp.error };
        }

        const destination = resp.data as INamespaceList;

        const data = {
          key: res.resource.key,
          payload: res.resource.payload,
          revision: destination.revision
        };

        resp = await baseQuery({
          url: `/${to.environmentKey}/namespaces/${to.namespaceKey}/resources`,
          method: 'POST',
          body: data
        });
        if (resp.error) {
          return { error: resp.error };
        }
        return { data: undefined };
      },
      invalidatesTags: (_result, _error, { to }) => [
        { type: 'Flag', id: to.environmentKey + '/' + to.namespaceKey }
      ]
    })
  })
});

export const {
  useListFlagsQuery,
  useGetFlagQuery,
  useCreateFlagMutation,
  useDeleteFlagMutation,
  useUpdateFlagMutation,
  useCopyFlagMutation
} = flagsApi;
