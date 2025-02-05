import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { SortingState } from '@tanstack/react-table';
import { RootState } from '~/store';
import { IFlag, IFlagBase, IFlagList } from '~/types/Flag';
import { IResourceListResponse, IResourceResponse } from '~/types/Resource';
import { IRollout } from '~/types/Rollout';
import { IRule } from '~/types/Rule';
import { IVariant, IVariantBase } from '~/types/Variant';
import { baseQuery } from '~/utils/redux-rtk';

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
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
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
        }
      ],
      transformResponse: (response: IResourceResponse<IFlag>): IFlag => {
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
        return {
          ...response.resource.payload,
          rollouts: response.resource.payload.rollouts?.map(
            (r: IRollout, i: number) => ({
              ...r,
              rank: i
            })
          ),
          rules: response.resource.payload.rules?.map(
            (r: IRule, i: number) => ({
              ...r,
              rank: i
            })
          ),
          variants: response.resource.payload.variants?.map(
            (v: IVariant, i: number) => ({
              ...v
            })
          )
        };
      }
    }),
    // create a new flag in the namespace
    createFlag: builder.mutation<
      IFlag,
      {
        environmentKey: string;
        namespaceKey: string;
        values: IFlagBase;
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
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, values }
      ) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Flag',
          id: environmentKey + '/' + namespaceKey + '/' + values.key
        }
      ]
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
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, flagKey }
      ) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Flag',
          id: environmentKey + '/' + namespaceKey + '/' + flagKey
        }
      ]
    }),
    // update the flag in the namespace
    updateFlag: builder.mutation<
      IFlag,
      {
        environmentKey: string;
        namespaceKey: string;
        flagKey: string;
        values: IFlag;
        revision?: string;
      }
    >({
      query({ environmentKey, namespaceKey, flagKey, values, revision }) {
        console.log('values', values);
        const payload = {
          '@type': 'flipt.core.Flag',
          defaultVariantId: values.defaultVariant?.id || null,
          metadata: values.metadata,
          ...values
        };
        console.log('payload', payload);
        delete payload.defaultVariant;
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources`,
          method: 'PUT',
          body: {
            '@type': 'flipt.core.Flag',
            key: flagKey,
            revision,
            payload
          }
        };
      },
      invalidatesTags: (
        _result,
        _error,
        { environmentKey, namespaceKey, flagKey }
      ) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        {
          type: 'Flag',
          id: environmentKey + '/' + namespaceKey + '/' + flagKey
        }
      ]
    }),
    // copy the flag from one namespace to another one
    copyFlag: builder.mutation<
      void,
      {
        environmentKey: string;
        from: { namespaceKey: string; flagKey: string };
        to: { namespaceKey: string; flagKey: string };
      }
    >({
      queryFn: async (
        { environmentKey, from, to },
        _api,
        _extraOptions,
        baseQuery
      ) => {
        let resp = await baseQuery({
          url: `/${environmentKey}/namespaces/${from.namespaceKey}/resources/flipt.core.Flag/${from.flagKey}`,
          method: 'GET'
        });
        if (resp.error) {
          return { error: resp.error };
        }

        const res = resp.data as {
          resource: { payload: IFlag; key: string };
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
        { type: 'Flag', id: environmentKey + '/' + to.namespaceKey }
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
