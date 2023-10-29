/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  createAsyncThunk,
  createSelector,
  createSlice
} from '@reduxjs/toolkit';
import { useNavigate } from 'react-router-dom';
import {
  copyFlag,
  createFlag,
  deleteFlag,
  getFlag,
  updateFlag
} from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { RootState } from '~/store';
import { IFlag, IFlagBase } from '~/types/Flag';
import { LoadingStatus } from '~/types/Meta';

const namespaceKey = 'namespace';

interface IFlagState {
  status: LoadingStatus;
  currentFlag: IFlag | null;
  error: string | undefined;
}

const initialState: IFlagState = {
  status: LoadingStatus.IDLE,
  currentFlag: null,
  error: undefined
};

export const flagsSlice = createSlice({
  name: 'flags',
  initialState,
  reducers: {},
  extraReducers(builder) {
    builder
      .addCase(fetchFlagAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(fetchFlagAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        state.currentFlag = action.payload;
      })
      .addCase(fetchFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(createFlagAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(createFlagAsync.fulfilled, (state, action) => {
        state.currentFlag = action.payload;
      })
      .addCase(createFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(updateFlagAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(updateFlagAsync.fulfilled, (state, action) => {
        state.currentFlag = action.payload;
      })
      .addCase(updateFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(deleteFlagAsync.fulfilled, (state, _action) => {
        state.currentFlag = null;
      })
      .addCase(deleteFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      });
  }
});

export const selectCurrentFlag = createSelector(
  [
    (state: RootState) => {
      if (state.flags.currentFlag) {
        return state.flags.currentFlag;
      }
      return { key: 'default', name: 'Default', description: '' } as IFlag;
    }
  ],
  (currentFlag) => currentFlag
);

type FlagIdentifier = {
  namespaceKey: string;
  key: string;
};

export const fetchFlagAsync = createAsyncThunk(
  'flags/fetchFlag',
  async (payload: FlagIdentifier) => {
    const { namespaceKey, key } = payload;
    return await getFlag(namespaceKey, key);
  }
);

export const copyFlagAsync = createAsyncThunk(
  'flags/copyFlag',
  async (payload: { from: FlagIdentifier; to: FlagIdentifier }) => {
    const { from, to } = payload;
    const response = await copyFlag(from, to);
    return response;
  }
);

export const createFlagAsync = createAsyncThunk(
  'flags/createFlag',
  async (payload: { namespaceKey: string; values: IFlagBase }) => {
    const { namespaceKey, values } = payload;
    const response = await createFlag(namespaceKey, values);
    return response;
  }
);

export const updateFlagAsync = createAsyncThunk(
  'flags/updateFlag',
  async (payload: { namespaceKey: string; key: string; values: IFlagBase }) => {
    const { namespaceKey, key, values } = payload;
    const response = await updateFlag(namespaceKey, key, values);
    return response;
  }
);

export const deleteFlagAsync = createAsyncThunk(
  'flags/deleteFlag',
  async (payload: FlagIdentifier) => {
    const { namespaceKey, key } = payload;
    const response = await deleteFlag(namespaceKey, key);
    return response;
  }
);

export default flagsSlice.reducer;
