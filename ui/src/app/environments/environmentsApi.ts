import { createSelector, createSlice } from '@reduxjs/toolkit';
import { createApi } from '@reduxjs/toolkit/query/react';

import { IBranchEnvironment, IEnvironment } from '~/types/Environment';
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
      const environment = action.payload;
      state.currentEnvironment = environment.key;
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
  endpoints: (builder) => ({
    listEnvironments: builder.query<{ environments: IEnvironment[] }, void>({
      query: () => '',
      providesTags: () => [{ type: 'Environment' }]
    }),
    listBranchEnvironments: builder.query<
      { branches: IBranchEnvironment[] },
      { baseEnvironmentKey: string }
    >({
      query: ({ baseEnvironmentKey }) => `/${baseEnvironmentKey}/branches`,
      providesTags: () => [{ type: 'BranchEnvironment' }]
    }),
    createBranchEnvironment: builder.mutation<
      IBranchEnvironment,
      { baseEnvironmentKey: string; environmentKey: string }
    >({
      query: ({ baseEnvironmentKey, environmentKey }) => ({
        url: `/${baseEnvironmentKey}/branches`,
        method: 'POST',
        body: {
          environmentKey
        }
      }),
      invalidatesTags: () => [
        { type: 'Environment' },
        { type: 'BranchEnvironment' }
      ]
    })
  })
});

export const {
  useListEnvironmentsQuery,
  useListBranchEnvironmentsQuery,
  useCreateBranchEnvironmentMutation
} = environmentsApi;

export const environmentsReducer = environmentsSlice.reducer;
