/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  createAsyncThunk,
  createSelector,
  createSlice
} from '@reduxjs/toolkit';
import {
  copyFlag,
  createFlag,
  deleteFlag,
  deleteVariant,
  getFlag,
  updateFlag
} from '~/data/api';
import { RootState } from '~/store';
import { IFlag, IFlagBase } from '~/types/Flag';
import { LoadingStatus } from '~/types/Meta';

const namespaceKey = 'namespace';

interface IFlagState {
  status: LoadingStatus;
  currentFlag: string | null;
  flags: { [key: string]: { [key: string]: IFlag } };
  error: string | undefined;
}

const initialState: IFlagState = {
  status: LoadingStatus.IDLE,
  currentFlag: null,
  flags: {},
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
        const [flag, namespaceKey] = [
          action.payload,
          action.meta.arg.namespaceKey
        ];
        state.currentFlag = flag.key;
        if (state.flags[namespaceKey] === undefined) {
          state.flags[namespaceKey] = {};
        }
        state.flags[namespaceKey][flag.key] = flag;
      })
      .addCase(fetchFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(createFlagAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(createFlagAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        const [flag, namespaceKey] = [
          action.payload,
          action.meta.arg.namespaceKey
        ];
        state.currentFlag = flag.key;
        if (state.flags[namespaceKey] === undefined) {
          state.flags[namespaceKey] = {};
        }
        state.flags[namespaceKey][flag.key] = flag;
      })
      .addCase(createFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(updateFlagAsync.pending, (state, _action) => {
        state.status = LoadingStatus.LOADING;
      })
      .addCase(updateFlagAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        const [flag, namespaceKey] = [
          action.payload,
          action.meta.arg.namespaceKey
        ];
        state.currentFlag = flag.key;
        if (state.flags[namespaceKey] === undefined) {
          state.flags[namespaceKey] = {};
        }
        state.flags[namespaceKey][flag.key] = flag;
      })
      .addCase(updateFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(deleteFlagAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        const flags = state.flags[action.meta.arg.namespaceKey];
        if (flags) {
          delete flags[action.meta.arg.key];
        }
        state.currentFlag = null;
      })
      .addCase(deleteFlagAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      })
      .addCase(deleteVariantAsync.fulfilled, (state, action) => {
        state.status = LoadingStatus.SUCCEEDED;
        const flags = state.flags[action.meta.arg.namespaceKey];
        if (flags) {
          const variants = flags[action.meta.arg.flagKey].variants || [];
          flags[action.meta.arg.flagKey].variants = variants.filter(
            (variant) => variant.id !== action.meta.arg.variantId
          );
        }
      })
      .addCase(deleteVariantAsync.rejected, (state, action) => {
        state.status = LoadingStatus.FAILED;
        state.error = action.error.message;
      });
  }
});

export const selectCurrentFlag = createSelector(
  [
    (state: RootState) => {
      const flags = state.flags.flags[state.namespaces.currentNamespace];
      if (flags && state.flags.currentFlag) {
        return flags[state.flags.currentFlag] || ({} as IFlag);
      }

      return {} as IFlag;
    }
  ],
  (currentFlag) => currentFlag
);

interface FlagIdentifier {
  namespaceKey: string;
  key: string;
}

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

interface VariantIdentifier {
  namespaceKey: string;
  flagKey: string;
  variantId: string;
}

export const deleteVariantAsync = createAsyncThunk(
  'flags/deleteVariant',
  async (payload: VariantIdentifier) => {
    const { namespaceKey, flagKey, variantId } = payload;
    const response = await deleteVariant(namespaceKey, flagKey, variantId);
    return response;
  }
);

export default flagsSlice.reducer;
