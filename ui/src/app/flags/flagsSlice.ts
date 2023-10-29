/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  createAsyncThunk,
  createSelector,
  createSlice
} from '@reduxjs/toolkit';
import { copyFlag, deleteFlag, getFlag } from '~/data/api';
import { RootState } from '~/store';
import { IFlag } from '~/types/Flag';
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
  'flags/asyncFlag',
  async (payload: { from: FlagIdentifier, to: FlagIdentifier }) => {
    const { from, to } = payload;
    await copyFlag(from, to);
    return to.key;
  }
);

export const deleteFlagAsync = createAsyncThunk(
  'flags/deleteFlag',
  async (payload: FlagIdentifier) => {
    const { namespaceKey, key } = payload;
    await deleteFlag(namespaceKey, key);
    return key;
  }
);

export default flagsSlice.reducer;
