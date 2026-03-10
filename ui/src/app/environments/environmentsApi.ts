import { createSelector, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';

import { IChange } from '~/types/Change';
import {
  IBranchEnvironment,
  IEnvironment,
  IEnvironmentProposal
} from '~/types/Environment';
import { LoadingStatus } from '~/types/Meta';

import { RootState } from '~/store';
import { baseQuery } from '~/utils/redux-rtk';

export const environmentKey = 'environment';

interface IEnvironmentsState {
  environments: { [key: string]: IEnvironment };
  status: LoadingStatus;
  currentEnvironment: string;
  error: string | undefined;
}

const initialState: IEnvironmentsState = {
  environments: {},
  status: LoadingStatus.IDLE,
  currentEnvironment: localStorage.getItem(environmentKey) || '',
  error: undefined
};

export const environmentsSlice = createSlice({
  name: 'environments',
  initialState,
  reducers: {
    currentEnvironmentChanged: (state, action) => {
      state.currentEnvironment = action.payload;
    },
    environmentsChanged: (state, action) => {
      const environments: { [key: string]: IEnvironment } = {};
      // First, build the environments map
      action.payload.environments.forEach((environment: IEnvironment) => {
        environments[environment.key] = environment;
      });
      state.environments = environments;
      state.status = LoadingStatus.SUCCEEDED;

      // If no current environment is set, or the current one doesn't exist anymore
      if (
        !state.currentEnvironment ||
        !environments[state.currentEnvironment]
      ) {
        // First, try to find the default environment
        const defaultEnv = action.payload.environments.find(
          (env: IEnvironment) => env.default === true
        );

        if (defaultEnv) {
          state.currentEnvironment = defaultEnv.name;
        } else if (action.payload.environments.length > 0) {
          // Only fallback to first environment if no default exists
          state.currentEnvironment = action.payload.environments[0].name;
        }
      }
    }
  }
});

export const { currentEnvironmentChanged, environmentsChanged } =
  environmentsSlice.actions;

// only base environments
export const selectEnvironments = createSelector(
  [(state: RootState) => state.environments.environments],
  (environments) => {
    return Object.entries(environments)
      .map(([_, value]) => value)
      .filter((env) => env.configuration?.base === undefined) as IEnvironment[]; // ignore branched environments
  }
);

// all environments, including branched environments
export const selectAllEnvironments = createSelector(
  [(state: RootState) => state.environments.environments],
  (environments) => {
    return Object.entries(environments).map(
      ([_, value]) => value
    ) as IEnvironment[];
  }
);

export const selectCurrentEnvironment = createSelector(
  [(state: RootState) => state.environments],
  (state) => {
    if (state.environments[state.currentEnvironment]) {
      return state.environments[state.currentEnvironment];
    }

    if (state.environments.default) {
      return state.environments.default;
    }

    const envs = Object.keys(state.environments);
    if (envs.length > 0) {
      return state.environments[envs[0]];
    }

    return { key: 'default', storage: '', directory: '' } as IEnvironment;
  }
);

export const environmentsApi = createApi({
  reducerPath: 'environments-api',
  baseQuery,
  tagTypes: ['Environment', 'BranchEnvironment'],
  refetchOnFocus: true,
  refetchOnReconnect: true,
  endpoints: (builder) => ({
    listEnvironments: builder.query<{ environments: IEnvironment[] }, void>({
      query: () => '',
      providesTags: () => [{ type: 'Environment' }]
    }),
    listBranchEnvironments: builder.query<
      { branches: IBranchEnvironment[] },
      { environmentKey: string }
    >({
      query: ({ environmentKey }) => `/${environmentKey}/branches`,
      providesTags: () => [{ type: 'BranchEnvironment' }]
    }),
    createBranchEnvironment: builder.mutation<
      IBranchEnvironment,
      { environmentKey: string; key: string }
    >({
      query: ({ environmentKey, key }) => ({
        url: `/${environmentKey}/branches`,
        method: 'POST',
        body: {
          key
        }
      }),
      invalidatesTags: () => [
        { type: 'Environment' },
        { type: 'BranchEnvironment' }
      ]
    }),
    deleteBranchEnvironment: builder.mutation<
      void,
      { environmentKey: string; key: string }
    >({
      query: ({ environmentKey, key }) => ({
        url: `/${environmentKey}/branches/${key}`,
        method: 'DELETE'
      }),
      invalidatesTags: () => [
        { type: 'Environment' },
        { type: 'BranchEnvironment' }
      ]
    }),
    listBranchEnvironmentChanges: builder.query<
      { changes: IChange[] },
      { environmentKey: string; key: string }
    >({
      query: ({ environmentKey, key }) =>
        `/${environmentKey}/branches/${key}/changes`
    }),
    proposeEnvironment: builder.mutation<
      IEnvironmentProposal,
      {
        environmentKey: string;
        key: string;
        title?: string;
        body?: string;
        draft?: boolean;
      }
    >({
      query: ({ environmentKey, key, title, body, draft }) => ({
        url: `${environmentKey}/branches/${key}`,
        method: 'POST',
        body: { title, body, draft }
      }),
      invalidatesTags: () => [{ type: 'BranchEnvironment' }]
    }),
    bulkApplyResources: builder.mutation<
      {
        results: Array<{
          namespaceKey: string;
          status: string;
          error?: string;
        }>;
        revision: string;
      },
      {
        environmentKey: string;
        namespaceKeys: string[];
        operation: string;
        typeUrl: string;
        key: string;
        payload?: unknown;
        onConflict?: string;
        revision: string;
      }
    >({
      query: ({
        environmentKey,
        namespaceKeys,
        operation,
        typeUrl,
        key,
        payload,
        onConflict,
        revision
      }) => ({
        url: `/${environmentKey}/resources/bulk`,
        method: 'POST',
        body: {
          namespace_keys: namespaceKeys,
          operation,
          type_url: typeUrl,
          key,
          payload,
          on_conflict: onConflict,
          revision
        }
      }),
      invalidatesTags: () => [{ type: 'Environment' }]
    }),
    copyNamespace: builder.mutation<
      {
        results: Array<{
          typeUrl: string;
          key: string;
          status: string;
          error?: string;
        }>;
        revision: string;
      },
      {
        environmentKey: string;
        namespaceKey: string;
        sourceEnvironmentKey: string;
        sourceNamespaceKey: string;
        onConflict?: string;
        revision: string;
      }
    >({
      query: ({
        environmentKey,
        namespaceKey,
        sourceEnvironmentKey,
        sourceNamespaceKey,
        onConflict,
        revision
      }) => ({
        url: `/${environmentKey}/namespaces/copy`,
        method: 'POST',
        body: {
          source_environment_key: sourceEnvironmentKey,
          source_namespace_key: sourceNamespaceKey,
          namespace_key: namespaceKey,
          on_conflict: onConflict,
          revision
        }
      }),
      invalidatesTags: () => [{ type: 'Environment' }]
    })
  })
});

export const {
  useListEnvironmentsQuery,
  useListBranchEnvironmentsQuery,
  useCreateBranchEnvironmentMutation,
  useDeleteBranchEnvironmentMutation,
  useListBranchEnvironmentChangesQuery,
  useProposeEnvironmentMutation,
  useBulkApplyResourcesMutation,
  useCopyNamespaceMutation
} = environmentsApi;

export const environmentsReducer = environmentsSlice.reducer;
