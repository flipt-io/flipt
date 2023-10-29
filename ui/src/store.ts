import {
  configureStore,
  createListenerMiddleware,
  isAnyOf
} from '@reduxjs/toolkit';

import { flagsSlice } from './app/flags/flagsSlice';
import { metaSlice } from './app/meta/metaSlice';
import { namespacesSlice } from './app/namespaces/namespacesSlice';
import {
  preferencesKey,
  preferencesSlice
} from './app/preferences/preferencesSlice';
import { IFlag } from './types/Flag';
import { LoadingStatus } from './types/Meta';

const listenerMiddleware = createListenerMiddleware();

const namespaceKey = 'namespace';

listenerMiddleware.startListening({
  matcher: isAnyOf(
    preferencesSlice.actions.themeChanged,
    preferencesSlice.actions.timezoneChanged
  ),
  effect: (_action, api) => {
    // save to local storage
    localStorage.setItem(
      preferencesKey,
      JSON.stringify((api.getState() as RootState).preferences)
    );
  }
});

listenerMiddleware.startListening({
  matcher: isAnyOf(namespacesSlice.actions.currentNamespaceChanged),
  effect: (_action, api) => {
    // save to local storage
    localStorage.setItem(
      namespaceKey,
      (api.getState() as RootState).namespaces.currentNamespace
    );
  }
});

const preferencesState = JSON.parse(
  localStorage.getItem(preferencesKey) || '{}'
);

const currentNamespace = localStorage.getItem(namespaceKey) || 'default';

export const store = configureStore({
  preloadedState: {
    preferences: preferencesState,
    namespaces: {
      namespaces: {},
      status: LoadingStatus.IDLE,
      currentNamespace,
      error: undefined
    },
    flags: {
      status: LoadingStatus.IDLE,
      flags: {},
      error: undefined
    }
  },
  reducer: {
    namespaces: namespacesSlice.reducer,
    preferences: preferencesSlice.reducer,
    flags: flagsSlice.reducer,
    meta: metaSlice.reducer
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().prepend(listenerMiddleware.middleware)
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
