/* eslint-disable @typescript-eslint/no-use-before-define */
import {
  ActionReducerMapBuilder,
  createAsyncThunk,
  createSelector,
  createSlice,
  SerializedError
} from '@reduxjs/toolkit';
import {
  copyFlag,
  createFlag,
  createVariant,
  deleteFlag,
  deleteVariant,
  getFlag,
  updateFlag,
  updateVariant
} from '~/data/api';
import { RootState } from '~/store';
import { IFlag, IFlagBase } from '~/types/Flag';
import { LoadingStatus } from '~/types/Meta';
import { IVariant, IVariantBase } from '~/types/Variant';

interface IFlagState {
  status: LoadingStatus;
  flags: { [key: string]: { [key: string]: IFlag } };
  error: string | undefined;
}

const initialState: IFlagState = {
  status: LoadingStatus.IDLE,
  flags: {},
  error: undefined
};

export const flagsSlice = createSlice({
  name: 'flags',
  initialState,
  reducers: {},
  extraReducers(builder) {
    [flagReducers, variantReducers].map((reducer) => reducer(builder));
  }
});

const flagReducers = (builder: ActionReducerMapBuilder<IFlagState>) => {
  builder
    .addCase(fetchFlagAsync.pending, setLoading)
    .addCase(fetchFlagAsync.fulfilled, setFlag)
    .addCase(fetchFlagAsync.rejected, setError)
    .addCase(createFlagAsync.pending, setLoading)
    .addCase(createFlagAsync.fulfilled, setFlag)
    .addCase(createFlagAsync.rejected, setError)
    .addCase(updateFlagAsync.pending, setLoading)
    .addCase(updateFlagAsync.fulfilled, setFlag)
    .addCase(updateFlagAsync.rejected, setError)
    .addCase(deleteFlagAsync.fulfilled, (state, action) => {
      state.status = LoadingStatus.SUCCEEDED;
      const flags = state.flags[action.meta.arg.namespaceKey];
      if (flags) {
        delete flags[action.meta.arg.key];
      }
    })
    .addCase(deleteFlagAsync.rejected, setError);
};

const variantReducers = (builder: ActionReducerMapBuilder<IFlagState>) => {
  builder
    .addCase(createVariantAsync.pending, setLoading)
    .addCase(
      createVariantAsync.fulfilled,
      (
        state: IFlagState,
        action: { payload: IVariant; meta: { arg: VariantValues } }
      ) => {
        state.status = LoadingStatus.SUCCEEDED;
        const flags = state.flags[action.meta.arg.namespaceKey];
        if (flags) {
          flags[action.meta.arg.flagKey].variants?.push(action.payload);
        }
      }
    )
    .addCase(createVariantAsync.rejected, setError)
    .addCase(updateVariantAsync.pending, setLoading)
    .addCase(
      updateVariantAsync.fulfilled,
      (
        state: IFlagState,
        action: { payload: IVariant; meta: { arg: VariantValues } }
      ) => {
        state.status = LoadingStatus.SUCCEEDED;
        const flags = state.flags[action.meta.arg.namespaceKey];
        if (flags) {
          const variants = flags[action.meta.arg.flagKey].variants || [];
          const idx = variants.findIndex(
            (v) => v.key == action.meta.arg.values.key
          );
          if (idx >= 0) {
            variants[idx] = action.payload;
          }
          flags[action.meta.arg.flagKey].variants = variants;
        }
      }
    )
    .addCase(updateVariantAsync.rejected, setError)
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
    .addCase(deleteVariantAsync.rejected, setError);
};

const setLoading = (state: IFlagState, _action: any) => {
  state.status = LoadingStatus.LOADING;
};

const setError = (
  state: IFlagState,
  action: { payload: any; error: SerializedError }
) => {
  state.status = LoadingStatus.FAILED;
  state.error = action.error.message;
};

const setFlag = (
  state: IFlagState,
  action: { payload: IFlag; meta: { arg: { namespaceKey: string } } }
) => {
  state.status = LoadingStatus.SUCCEEDED;
  const [flag, namespaceKey] = [action.payload, action.meta.arg.namespaceKey];
  if (state.flags[namespaceKey] === undefined) {
    state.flags[namespaceKey] = {};
  }
  state.flags[namespaceKey][flag.key] = flag;
};

const selectNamespaceFlags = (state: RootState) =>
  state.flags.flags[state.namespaces.currentNamespace];

export const selectFlag = createSelector(
  [selectNamespaceFlags, (state: RootState, key: string) => key],
  (flags, key) => {
    if (flags === undefined) {
      return {} as IFlag;
    }
    return flags[key];
  }
);

interface FlagIdentifier {
  namespaceKey: string;
  key: string;
}

export const fetchFlagAsync = createAsyncThunk(
  'flags/fetchFlag',
  async (payload: FlagIdentifier) => {
    const { namespaceKey, key } = payload;
    const response = await getFlag(namespaceKey, key);
    return response;
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

interface VariantValues {
  namespaceKey: string;
  flagKey: string;
  values: IVariantBase;
}

export const createVariantAsync = createAsyncThunk(
  'flags/createVariant',
  async (payload: VariantValues) => {
    const { namespaceKey, flagKey, values } = payload;
    const response = await createVariant(namespaceKey, flagKey, values);
    return response;
  }
);

export const updateVariantAsync = createAsyncThunk(
  'flags/updateVariant',
  async (payload: VariantIdentifier & VariantValues) => {
    const { namespaceKey, flagKey, variantId, values } = payload;
    const response = await updateVariant(
      namespaceKey,
      flagKey,
      variantId,
      values
    );
    return response;
  }
);

export default flagsSlice.reducer;
