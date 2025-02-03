/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { getMeta } from '~/data/api';
import { IInfo, LoadingStatus } from '~/types/Meta';

interface IMetaSlice {
  info: IInfo & { status: LoadingStatus };
}

const initialState: IMetaSlice = {
  info: {
    version: '0.0.0',
    latestVersion: '0.0.0',
    latestVersionURL: '',
    commit: '',
    buildDate: '',
    goVersion: '',
    updateAvailable: false,
    isRelease: false,
    analyticsEnabled: false,
    status: LoadingStatus.IDLE
  }
};

export const metaSlice = createSlice({
  name: 'meta',
  initialState,
  reducers: {},
  extraReducers(builder) {
    builder
      .addCase(fetchInfoAsync.pending, (state, _action) => {
        state.info.status = LoadingStatus.LOADING;
      })
      .addCase(fetchInfoAsync.fulfilled, (state, action) => {
        state.info = action.payload;
        state.info.status = LoadingStatus.SUCCEEDED;
      });
  }
});

export const selectInfo = (state: { meta: IMetaSlice }) => state.meta.info;

export const fetchInfoAsync = createAsyncThunk('meta/fetchInfo', async () => {
  const response = await getMeta();
  return response;
});

export default metaSlice.reducer;
