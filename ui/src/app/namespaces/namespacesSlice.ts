/* eslint-disable @typescript-eslint/no-use-before-define */
import { createSelector, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { RootState } from '~/store';
import { LoadingStatus } from '~/types/Meta';
import { INamespace, INamespaceBase, INamespaceList } from '~/types/Namespace';
import { baseQuery } from '~/utils/redux-rtk';

export const namespaceKey = 'namespace';

interface INamespacesState {
  namespaces: { [key: string]: INamespace };
  status: LoadingStatus;
  currentNamespace: string;
  error: string | undefined;
}

const initialState: INamespacesState = {
  namespaces: {},
  status: LoadingStatus.IDLE,
  currentNamespace: localStorage.getItem(namespaceKey) || 'default',
  error: undefined
};

export const namespacesSlice = createSlice({
  name: 'namespaces',
  initialState,
  reducers: {
    currentNamespaceChanged: (state, action) => {
      const namespace = action.payload;
      state.currentNamespace = namespace.key;
    },
    namespacesChanged: (state, action) => {
      const namespaces: { [key: string]: INamespace } = {};
      action.payload.namespaces.forEach((namespace: INamespace) => {
        namespaces[namespace.key] = namespace;
      });
      state.namespaces = namespaces;
      state.status = LoadingStatus.SUCCEEDED;
    }
  }
});

export const { currentNamespaceChanged } = namespacesSlice.actions;

export const selectNamespaces = createSelector(
  [(state: RootState) => state.namespaces.namespaces],
  (namespaces) => {
    return Object.entries(namespaces).map(
      ([_, value]) => value
    ) as INamespace[];
  }
);

export const selectCurrentNamespace = createSelector(
  [(state: RootState) => state.namespaces],
  (namespaces) =>
    namespaces.namespaces[namespaces.currentNamespace] ||
    ({ key: 'default', name: 'Default', description: '' } as INamespace)
);

export const namespaceApi = createApi({
  reducerPath: 'namespaces-api',
  baseQuery,
  tagTypes: ['Namespace'],
  endpoints: (builder) => ({
    // get list of namespaces
    listNamespaces: builder.query<INamespaceList, void>({
      query: () => '/namespaces',
      providesTags: () => [{ type: 'Namespace' }]
    }),
    // create the namespace
    createNamespace: builder.mutation<INamespace, INamespaceBase>({
      query(values) {
        return {
          url: '/namespaces',
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: () => [{ type: 'Namespace' }]
    }),
    // update the namespace
    updateNamespace: builder.mutation<
      INamespace,
      { key: string; values: INamespaceBase }
    >({
      query({ key, values }) {
        return {
          url: `/namespaces/${key}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: () => [{ type: 'Namespace' }]
    }),
    // delete the namespace
    deleteNamespace: builder.mutation<void, string>({
      query(namespaceKey) {
        return {
          url: `/namespaces/${namespaceKey}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: () => [{ type: 'Namespace' }]
    })
  })
});

export const {
  useListNamespacesQuery,
  useCreateNamespaceMutation,
  useDeleteNamespaceMutation,
  useUpdateNamespaceMutation
} = namespaceApi;

export default namespacesSlice.reducer;
