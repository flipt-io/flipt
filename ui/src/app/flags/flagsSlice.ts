/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  createAsyncThunk,
  createSelector,
  createSlice
} from '@reduxjs/toolkit';
import { getFlag } from '~/data/api';
import { RootState } from '~/store';
import { IFlag } from '~/types/Flag';
import { LoadingStatus } from '~/types/Meta';

const namespaceKey = 'namespace';

interface IFlagState {
  status: LoadingStatus;
  currentFlag: IFlag;
  error: string | undefined;
}

const initialState: IFlagState = {
  status: LoadingStatus.IDLE,
  currentFlag: {} as IFlag,
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

export const fetchFlagAsync = createAsyncThunk(
  'flags/fetchFlag',
  async (payload: { namespace: string; key: string }) => {
    const { namespace, key } = payload;
    return await getFlag(namespace, key);
  }
);

export default flagsSlice.reducer;
