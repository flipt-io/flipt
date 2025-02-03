import { createApi } from '@reduxjs/toolkit/query/react';
import { createSlice } from '@reduxjs/toolkit';
import { IEnvironment } from '~/types/Environment';
import { baseQuery } from '~/utils/redux-rtk';
import { LoadingStatus } from '~/types/Meta';

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
    setCurrentEnvironment: (state, action) => {
      state.currentEnvironment = action.payload;
    }
  }
});

export const { setCurrentEnvironment } = environmentsSlice.actions;

export const environmentsApi = createApi({
  reducerPath: 'environments-api',
  baseQuery: baseQuery,
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
