/* eslint-disable @typescript-eslint/no-use-before-define */
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { getConfig, getInfo } from '~/data/api';
import { IConfig, IInfo } from '~/types/Meta';

interface IMetaSlice {
  info: IInfo;
  config: IConfig;
  readonly: boolean;
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
    db: {
      url: ''
    },
    authentication: {
      required: false
    }
  },
  readonly: false
};

export const metaSlice = createSlice({
  name: 'meta',
  initialState,
  reducers: {
    setInfo: (state, action) => {
      state.info = action.payload;
    },
    setReadonly: (state, action) => {
      state.readonly = action.payload;
    }
  },
  extraReducers(builder) {
    builder
      .addCase(fetchInfoAsync.fulfilled, (state, action) => {
        state.info = action.payload;
      })
      .addCase(fetchConfigAsync.fulfilled, (state, action) => {
        state.config = action.payload;
      });
  }
});

export const { setInfo } = metaSlice.actions;

export const selectInfo = (state: { meta: IMetaSlice }) => state.meta.info;
export const selectReadonly = (state: { meta: IMetaSlice }) =>
  state.meta.readonly;

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
