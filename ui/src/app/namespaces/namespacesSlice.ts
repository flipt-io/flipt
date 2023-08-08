/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  createAsyncThunk,
  createSelector,
  createSlice
} from '@reduxjs/toolkit';
import {
  createNamespace,
  deleteNamespace,
  listNamespaces,
  updateNamespace
} from '~/data/api';
import { RootState } from '~/store';
import { LoadingStatus } from '~/types/Meta';
import { INamespace, INamespaceBase } from '~/types/Namespace';

const namespaceKey = 'namespace';

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
    }
  },
  extraReducers(builder) {
    builder
      .addCase(fetchNamespacesAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(fetchNamespacesAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        state.namespaces = action.payload;
      })
      .addCase(fetchNamespacesAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(createNamespaceAsync.fulfilled, (state, action) => {
        const namespace = action.payload;
        state.namespaces[namespace.key] = namespace;
      })
      .addCase(updateNamespaceAsync.fulfilled, (state, action) => {
        const namespace = action.payload;
        state.namespaces[namespace.key] = namespace;
      })
      .addCase(deleteNamespaceAsync.fulfilled, (state, action) => {
        const key = action.payload;
        delete state.namespaces[key];
        // if the current namespace is the one being deleted, set the current namespace to the first one
        if (state.currentNamespace === key) {
          state.currentNamespace =
            state.namespaces[Object.keys(state.namespaces)[0]].key;
        }
      });
  }
});

export const { currentNamespaceChanged } = namespacesSlice.actions;

export const selectNamespaces = createSelector(
  [
    (state: RootState) =>
      Object.entries(state.namespaces.namespaces).map(
        ([_, value]) => value
      ) as INamespace[]
  ],
  (namespaces) => namespaces
);

export const selectCurrentNamespace = createSelector(
  [
    (state: RootState) => {
      if (state.namespaces.namespaces[state.namespaces.currentNamespace]) {
        return state.namespaces.namespaces[state.namespaces.currentNamespace];
      }
      return { key: 'default', name: 'Default', description: '' } as INamespace;
    }
  ],
  (currentNamespace) => currentNamespace
);

export const fetchNamespacesAsync = createAsyncThunk(
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

export const createNamespaceAsync = createAsyncThunk(
  'namespaces/createNamespace',
  async (namespace: INamespaceBase) => {
    const response = await createNamespace(namespace);
    return response;
  }
);

export const updateNamespaceAsync = createAsyncThunk(
  'namespaces/updateNamespace',
  async (payload: { key: string; namespace: INamespaceBase }) => {
    const { key, namespace } = payload;
    const response = await updateNamespace(key, namespace);
    return response;
  }
);

export const deleteNamespaceAsync = createAsyncThunk(
  'namespaces/deleteNamespace',
  async (key: string) => {
    await deleteNamespace(key);
    return key;
  }
);

export default namespacesSlice.reducer;
