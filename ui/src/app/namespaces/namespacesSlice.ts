/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { listNamespaces } from '~/data/api';
import { RootState } from '~/store';
import { INamespace } from '~/types/Namespace';

interface INamespacesState {
  namespaces: INamespace[];
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  selectedNamespace: INamespace | null;
  error: string | undefined;
}

const initialState: INamespacesState = {
  namespaces: [],
  status: 'idle',
  selectedNamespace: null,
  error: undefined
};

export const namespacesSlice = createSlice({
  name: 'namespaces',
  initialState,
  reducers: {
    namespaceCreated: (state, action) => {
      state.namespaces.push(action.payload);
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

export const { namespaceCreated } = namespacesSlice.actions;

export const selectNamespaces = (state: RootState) =>
  state.namespaces.namespaces;
export const selectNamespace = (state: RootState, namespaceKey: string) => {
  return state.namespaces.namespaces.find(
    (n: INamespace) => n.key === namespaceKey
  );
};

export const fetchNamespaces = createAsyncThunk(
  'namespaces/fetchNamespaces',
  async () => {
    const response = await listNamespaces();
    return response.namespaces;
  }
);

export default namespacesSlice.reducer;
