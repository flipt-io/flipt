/* eslint-disable @typescript-eslint/no-use-before-define */
import { createSelector, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';

import { LoadingStatus } from '~/types/Meta';
import { INamespace, INamespaceList } from '~/types/Namespace';

import { RootState } from '~/store';
import { baseQuery } from '~/utils/redux-rtk';

export const namespaceKey = 'namespace';

interface INamespacesState {
  namespaces: { [key: string]: INamespace };
  status: LoadingStatus;
  currentNamespace: string;
  revision: string;
  error: string | undefined;
}

const initialState: INamespacesState = {
  namespaces: {},
  status: LoadingStatus.IDLE,
  currentNamespace: localStorage.getItem(namespaceKey) || 'default',
  revision: '',
  error: undefined
};

export const namespacesSlice = createSlice({
  name: 'namespaces',
  initialState,
  reducers: {
    currentNamespaceChanged: (state, action) => {
      state.currentNamespace = action.payload;
    },
    namespacesChanged: (state, action) => {
      const namespaces: { [key: string]: INamespace } = {};
      action.payload.items.forEach((namespace: INamespace) => {
        namespaces[namespace.key] = namespace;
      });
      state.namespaces = namespaces;
      state.status = LoadingStatus.SUCCEEDED;
    },
    revisionChanged: (state, action) => {
      state.revision = action.payload;
    }
  }
});

export const { currentNamespaceChanged, revisionChanged } =
  namespacesSlice.actions;

export const selectRevision = (state: RootState) => state.namespaces.revision;

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
  (state) => {
    if (state.namespaces[state.currentNamespace]) {
      return state.namespaces[state.currentNamespace];
    }

    if (state.namespaces.default) {
      return state.namespaces.default;
    }

    const ns = Object.keys(state.namespaces);
    if (ns.length > 0) {
      return state.namespaces[ns[0]];
    }

    return { key: 'default', name: 'Default', description: '' } as INamespace;
  }
);

export const namespaceApi = createApi({
  reducerPath: 'namespaces-api',
  baseQuery,
  tagTypes: ['Namespace', 'Flag', 'Segment'],
  endpoints: (builder) => ({
    // get list of namespaces
    listNamespaces: builder.query<INamespaceList, { environmentKey: string }>({
      query: ({ environmentKey }) => `/${environmentKey}/namespaces`,
      providesTags: () => [{ type: 'Namespace' }],
      transformResponse: (response: INamespaceList): INamespaceList => {
        return response;
      }
    }),
    // create the namespace
    createNamespace: builder.mutation<
      INamespace,
      { environmentKey: string; values: INamespace; revision: string }
    >({
      query: ({ environmentKey, values, revision }) => ({
        url: `/${environmentKey}/namespaces`,
        method: 'POST',
        body: {
          ...values,
          revision
        }
      }),
      invalidatesTags: () => [
        { type: 'Namespace' },
        { type: 'Flag' },
        { type: 'Segment' }
      ]
    }),
    // update the namespace
    updateNamespace: builder.mutation<
      INamespace,
      { environmentKey: string; values: INamespace; revision: string }
    >({
      query: ({ environmentKey, values, revision }) => ({
        url: `/${environmentKey}/namespaces`,
        method: 'PUT',
        body: {
          ...values,
          revision
        }
      }),
      invalidatesTags: () => [
        { type: 'Namespace' },
        { type: 'Flag' },
        { type: 'Segment' }
      ]
    }),
    // delete the namespace
    deleteNamespace: builder.mutation<
      void,
      { environmentKey: string; namespaceKey: string; revision: string }
    >({
      query: ({ environmentKey, namespaceKey, revision }) => ({
        url: `/${environmentKey}/namespaces/${namespaceKey}?revision=${revision}`,
        method: 'DELETE'
      }),
      invalidatesTags: () => [
        { type: 'Namespace' },
        { type: 'Flag' },
        { type: 'Segment' }
      ]
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
