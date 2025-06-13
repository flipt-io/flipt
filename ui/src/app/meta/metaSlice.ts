/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import { IInfo, LoadingStatus, Product } from '~/types/Meta';

import { getMeta } from '~/data/api';

interface IMetaSlice {
  info: IInfo & { status: LoadingStatus };
}

const initialState: IMetaSlice = {
  info: {
    build: {
      version: '0.0.0',
      latestVersion: '0.0.0',
      latestVersionURL: '',
      commit: '',
      buildDate: '',
      updateAvailable: false,
      isRelease: false
    },
    product: Product.OSS,
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
