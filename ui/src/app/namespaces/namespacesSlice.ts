/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { listNamespaces } from '~/data/api';
import { RootState } from '~/store';
import { INamespace } from '~/types/Namespace';

interface INamespacesState {
  namespaces: { [key: string]: INamespace };
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  currentNamespace: string;
  error: string | undefined;
}

const initialState: INamespacesState = {
  namespaces: {},
  status: 'idle',
  currentNamespace: 'default',
  error: undefined
};

export const namespacesSlice = createSlice({
  name: 'namespaces',
  initialState,
  reducers: {
    namespaceCreated: (state, action) => {
      const namespace = action.payload;
      state.namespaces[namespace.key] = action.payload;
    },
    currentNamespaceChanged: (state, action) => {
      const namespace = action.payload;
      state.currentNamespace = namespace.key;
    }
  },
  extraReducers(builder) {
    builder
      .addCase(fetchNamespaces.pending, (state, _action) => {
        state.status = 'loading';
      })
      .addCase(fetchNamespaces.fulfilled, (state, action) => {
        state.status = 'succeeded';
        state.namespaces = action.payload;
      })
      .addCase(fetchNamespaces.rejected, (state, action) => {
        state.status = 'failed';
        state.error = action.error.message;
      });
  }
});

export const { namespaceCreated, currentNamespaceChanged } =
  namespacesSlice.actions;

export const selectNamespaces = (state: RootState) =>
  Object.entries(state.namespaces.namespaces).map(
    ([_, value]) => value
  ) as INamespace[];

export const currentNamespace = (state: RootState) =>
  state.namespaces.namespaces[state.namespaces.currentNamespace];

export const fetchNamespaces = createAsyncThunk(
  'namespaces/fetchNamespaces',
  async () => {
    const response = await listNamespaces();
    const namespaces: { [key: string]: INamespace } = {};
    response.namespaces.forEach((namespace: INamespace) => {
      namespaces[namespace.key] = namespace;
    });
    return namespaces;
  }
);

export default namespacesSlice.reducer;
