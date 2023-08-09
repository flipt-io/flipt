/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { getConfig, getInfo } from '~/data/api';
import { IConfig, IInfo, StorageType } from '~/types/Meta';

interface IMetaSlice {
  info: IInfo;
  config: IConfig;
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
    isRelease: false
  },
  config: {
    storage: {
      type: StorageType.DATABASE,
      readOnly: false
    }
  }
};

export const metaSlice = createSlice({
  name: 'meta',
  initialState,
  reducers: {},
  extraReducers(builder) {
    builder
      .addCase(fetchInfoAsync.fulfilled, (state, action) => {
        state.info = action.payload;
      })
      .addCase(fetchConfigAsync.fulfilled, (state, action) => {
        state.config = action.payload;
        if (action.payload.storage?.readOnly === undefined) {
          state.config.storage.readOnly =
            action.payload.storage?.type &&
            action.payload.storage?.type !== StorageType.DATABASE;
        }
      });
  }
});

export const selectInfo = (state: { meta: IMetaSlice }) => state.meta.info;
export const selectConfig = (state: { meta: IMetaSlice }) => state.meta.config;
export const selectReadonly = (state: { meta: IMetaSlice }) =>
  state.meta.config.storage.readOnly;

export const fetchInfoAsync = createAsyncThunk('meta/fetchInfo', async () => {
  const response = await getInfo();
  return response;
});

export const fetchConfigAsync = createAsyncThunk(
  'meta/fetchConfig',
  async () => {
    const response = await getConfig();
    return response;
  }
);

export default metaSlice.reducer;
