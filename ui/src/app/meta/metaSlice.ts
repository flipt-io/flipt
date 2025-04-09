/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { getInfo } from '~/data/api';
import { IConfig, IInfo, LoadingStatus, StorageType } from '~/types/Meta';
import { Theme } from '~/types/Preferences';

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
    status: LoadingStatus.IDLE,
    storage: {
      type: StorageType.DATABASE,
      readOnly: false
    },
    ui: {
      defaultTheme: Theme.SYSTEM,
      topbar: { color: '' }
    },
    analyticsEnabled: false
  }
};

export const metaSlice = createSlice({
  name: 'meta',
  initialState,
  reducers: {},
  extraReducers(builder) {
    builder
      .addCase(fetchInfoAsync.pending, (state, _action) => {
        state.config.status = LoadingStatus.LOADING;
      })
      .addCase(fetchInfoAsync.fulfilled, (state, action) => {
        state.info = action.payload;
        state.config.status = LoadingStatus.SUCCEEDED;
        state.config.analyticsEnabled =
          action.payload.analytics?.enabled || false;
        if (action.payload.storage !== undefined) {
          state.config.storage.type = action.payload.storage?.type;
          state.config.storage.git = action.payload.storage?.metadata;
          state.config.storage.readOnly = action.payload.storage?.readOnly;
        }
        state.config.ui = {
          defaultTheme: action.payload.ui?.theme || Theme.SYSTEM,
          topbar: {
            color: action.payload.ui?.topbarColor || ''
          }
        };
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

export default metaSlice.reducer;
