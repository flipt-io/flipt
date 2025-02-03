import { createApi } from '@reduxjs/toolkit/query/react';
import { createSelector, createSlice } from '@reduxjs/toolkit';
import { IEnvironment } from '~/types/Environment';
import { baseQuery } from '~/utils/redux-rtk';
import { LoadingStatus } from '~/types/Meta';
import { RootState } from '~/store';

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
  currentEnvironment: localStorage.getItem(environmentKey) || 'default',
  error: undefined
};

export const environmentsSlice = createSlice({
  name: 'environments',
  initialState,
  reducers: {
    currentEnvironmentChanged: (state, action) => {
      const environment = action.payload;
      state.currentEnvironment = environment.name;
    },
    environmentsChanged: (state, action) => {
      const environments: { [key: string]: IEnvironment } = {};
      action.payload.environments.forEach((environment: IEnvironment) => {
        environments[environment.name] = environment;
      });
      state.environments = environments;
      state.status = LoadingStatus.SUCCEEDED;
    }
  }
});

export const { currentEnvironmentChanged } = environmentsSlice.actions;

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

    return { name: 'default', storage: '', directory: '' } as IEnvironment;
  }
);

export const environmentsApi = createApi({
  reducerPath: 'environments-api',
  baseQuery,
  tagTypes: ['Environment'],
  endpoints: (builder) => ({
    listEnvironments: builder.query<{ environments: IEnvironment[] }, void>({
      query: () => '',
      providesTags: (result, _error) =>
        result?.environments.map(({ name }) => ({
          type: 'Environment' as const,
          id: name
        })) || []
    })
  })
});

export const { useListEnvironmentsQuery } = environmentsApi;

export const environmentsReducer = environmentsSlice.reducer;
