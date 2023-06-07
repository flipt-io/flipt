import {
  configureStore,
  createListenerMiddleware,
  isAnyOf
} from '@reduxjs/toolkit';

import { metaSlice } from './app/meta/metaSlice';
import { namespacesSlice } from './app/namespaces/namespacesSlice';
import { preferencesSlice } from './app/preferences/preferencesSlice';

const listenerMiddleware = createListenerMiddleware();

const preferencesKey = 'preferences';

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

const preferencesState = JSON.parse(
  localStorage.getItem(preferencesKey) || '{}'
);

export const store = configureStore({
  preloadedState: {
    preferences: preferencesState
  },
  reducer: {
    namespaces: namespacesSlice.reducer,
    preferences: preferencesSlice.reducer,
    meta: metaSlice.reducer
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().prepend(listenerMiddleware.middleware)
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
