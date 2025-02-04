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
      transformResponse: (response: IResourceListResponse<IFlag>): IFlagList => {
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
        return {
          flags: response.resources.map(({ payload }) => payload)
        } as IFlagList;
      }
    }),
    // get flag in this namespace
    getFlag: builder.query<IFlag, { environmentKey: string; namespaceKey: string; flagKey: string }>({
      query: ({ environmentKey, namespaceKey, flagKey }) =>
        `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag/${flagKey}`,
      providesTags: (_result, _error, { environmentKey, namespaceKey, flagKey }) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey + '/' + flagKey }
      ],
      transformResponse: (response: IResourceResponse<IFlag>): IFlag => {
        if (response.revision) {
          localStorage.setItem('revision', response.revision);
        }
        return {
          ...response.resource.payload,
          rollouts: response.resource.payload.rollouts?.map((r: IRollout, i: number) => ({
            ...r,
            rank: i
          })),
          rules: response.resource.payload.rules?.map((r: IRule, i: number) => ({
            ...r,
            rank: i
          })),
          variants: response.resource.payload.variants?.map((v: IVariant, i: number) => ({
            ...v
          }))
        };
      }
    }),
    // create a new flag in the namespace
    createFlag: builder.mutation<
      IFlag,
      { environmentKey: string; namespaceKey: string; values: IFlagBase, revision?: string }
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
      invalidatesTags: (_result, _error, { environmentKey, namespaceKey, values }) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        { type: 'Flag', id: environmentKey + '/' + namespaceKey + '/' + values.key }
      ]
    }),
    // delete the flag from the namespace
    deleteFlag: builder.mutation<
      void,
      { environmentKey: string; namespaceKey: string; flagKey: string }
    >({
      query({ environmentKey, namespaceKey, flagKey }) {
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag/${flagKey}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { environmentKey, namespaceKey, flagKey }) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        { type: 'Flag', id: environmentKey + '/' + namespaceKey + '/' + flagKey }
      ]
    }),
    // update the flag in the namespace
    updateFlag: builder.mutation<
      IFlag,
      { environmentKey: string; namespaceKey: string; flagKey: string; values: IFlagBase, revision?: string }
    >({
      query({ environmentKey, namespaceKey, flagKey, values, revision }) {
        const payload = {
          '@type': 'flipt.core.Flag',
          defaultVariantId: values.defaultVariant?.id || null,
          metadata: values.metadata || undefined,
          ...values
        };
        delete payload.defaultVariant;
        return {
          url: `/${environmentKey}/namespaces/${namespaceKey}/resources/flipt.core.Flag/${flagKey}`,
          method: 'PUT',
          body: {
            key: flagKey,
            revision,
            payload
          }
        };
      },
      invalidatesTags: (_result, _error, { environmentKey,  namespaceKey, flagKey }) => [
        { type: 'Flag', id: environmentKey + '/' + namespaceKey },
        { type: 'Flag', id: environmentKey + '/' + namespaceKey + '/' + flagKey }
      ]
    }),
    // copy the flag from one namespace to another one
    copyFlag: builder.mutation<
      void,
      {
        from: { namespaceKey: string; flagKey: string };
        to: { namespaceKey: string; flagKey: string };
      }
    >({
      queryFn: async ({ from, to }, _api, _extraOptions, baseQuery) => {
        let resp = await baseQuery({
          url: `/namespaces/${from.namespaceKey}/flags/${from.flagKey}`,
          method: 'get'
        });
        if (resp.error) {
          return { error: resp.error };
        }
        let data = resp.data as IFlag;

        if (to.flagKey) {
          data.key = to.flagKey;
        }
        // first create the flag
        resp = await baseQuery({
          url: `/namespaces/${to.namespaceKey}/flags`,
          method: 'POST',
          body: data
        });
        if (resp.error) {
          return { error: resp.error };
        }
        // then copy the variants
        const variants = data.variants || [];
        for (let variant of variants) {
          resp = await baseQuery({
            url: `/namespaces/${to.namespaceKey}/flags/${to.flagKey}/variants`,
            method: 'POST',
            body: variant
          });
          if (resp.error) {
            return { error: resp.error };
          }
        }

        return { data: undefined };
      },
      invalidatesTags: (_result, _error, { to }) => [
        { type: 'Flag', id: to.namespaceKey }
      ]
    }),

    // create the flag variant in the namespace
    createVariant: builder.mutation<
      void,
      { namespaceKey: string; flagKey: string; values: IVariantBase }
    >({
      query({ namespaceKey, flagKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/variants`,
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, flagKey }) => [
        { type: 'Flag', id: namespaceKey + '/' + flagKey }
      ]
    }),
    // update the flag variant in the namespace
    updateVariant: builder.mutation<
      void,
      {
        namespaceKey: string;
        flagKey: string;
        variantId: string;
        values: IVariantBase;
      }
    >({
      query({ namespaceKey, flagKey, variantId, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/variants/${variantId}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, flagKey }) => [
        { type: 'Flag', id: namespaceKey + '/' + flagKey }
      ]
    }),
    // delete the flag variant in the namespace
    deleteVariant: builder.mutation<
      void,
      { namespaceKey: string; flagKey: string; variantId: string }
    >({
      query({ namespaceKey, flagKey, variantId }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/variants/${variantId}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, flagKey }) => [
        { type: 'Flag', id: namespaceKey + '/' + flagKey }
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
  useCopyFlagMutation,
  useCreateVariantMutation,
  useUpdateVariantMutation,
  useDeleteVariantMutation
} = flagsApi;
