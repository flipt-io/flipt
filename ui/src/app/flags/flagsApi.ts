import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { SortingState } from '@tanstack/react-table';
import { RootState } from '~/store';
import { IFlag, IFlagBase, IFlagList } from '~/types/Flag';
import { IVariantBase } from '~/types/Variant';
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
    listFlags: builder.query<IFlagList, string>({
      query: (namespaceKey) => `/namespaces/${namespaceKey}/flags`,
      providesTags: (result, _error, namespaceKey) =>
        result
          ? [
              ...result.flags.map(({ key }) => ({
                type: 'Flag' as const,
                id: namespaceKey + '/' + key
              })),
              { type: 'Flag', id: namespaceKey }
            ]
          : [{ type: 'Flag', id: namespaceKey }]
    }),
    // get flag in this namespace
    getFlag: builder.query<IFlag, { namespaceKey: string; flagKey: string }>({
      query: ({ namespaceKey, flagKey }) =>
        `/namespaces/${namespaceKey}/flags/${flagKey}`,
      providesTags: (_result, _error, { namespaceKey, flagKey }) => [
        { type: 'Flag', id: namespaceKey + '/' + flagKey }
      ]
    }),
    // create a new flag in the namespace
    createFlag: builder.mutation<
      IFlag,
      { namespaceKey: string; values: IFlagBase }
    >({
      query({ namespaceKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags`,
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, values }) => [
        { type: 'Flag', id: namespaceKey },
        { type: 'Flag', id: namespaceKey + '/' + values.key }
      ]
    }),
    // delete the flag from the namespace
    deleteFlag: builder.mutation<
      void,
      { namespaceKey: string; flagKey: string }
    >({
      query({ namespaceKey, flagKey }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey }) => [
        { type: 'Flag', id: namespaceKey }
      ]
    }),
    // update the flag in the namespace
    updateFlag: builder.mutation<
      IFlag,
      { namespaceKey: string; flagKey: string; values: IFlagBase }
    >({
      query({ namespaceKey, flagKey, values }) {
        // create new object 'values' to remap defaultVariant to defaultVariantId
        const update = {
          defaultVariantId: values.defaultVariant?.id || null,
          ...values
        };
        delete update.defaultVariant;
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}`,
          method: 'PUT',
          body: update
        };
      },
      invalidatesTags: (_result, _error, { namespaceKey, flagKey }) => [
        { type: 'Flag', id: namespaceKey },
        { type: 'Flag', id: namespaceKey + '/' + flagKey }
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
