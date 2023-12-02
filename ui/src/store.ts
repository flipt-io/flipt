import {
  configureStore,
  createListenerMiddleware,
  isAnyOf
} from '@reduxjs/toolkit';

import {
  namespaceApi,
  namespacesSlice
} from '~/app/namespaces/namespacesSlice';
import { flagsApi } from './app/flags/flagsApi';
import { rolloutsApi } from './app/flags/rolloutsApi';
import { rulesApi } from './app/flags/rulesApi';
import { metaSlice } from './app/meta/metaSlice';
import {
  preferencesKey,
  preferencesSlice
} from './app/preferences/preferencesSlice';
import { segmentsApi } from './app/segments/segmentsApi';
import { tokensApi } from './app/tokens/tokensApi';
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
/*
 * It could be anti-pattern but it feels like the right thing to do.
 * The namespacesSlice holds the namespaces globally and doesn't refetch them
 * from server as often as it could be with `useListNamespacesQuery`.
 * Each time the the app is loaded or namespaces changed by user this
 * listener will propagate the latest namespaces to the namespacesSlice.
 */
listenerMiddleware.startListening({
  matcher: namespaceApi.endpoints.listNamespaces.matchFulfilled,
  effect: (action, api) => {
    api.dispatch(namespacesSlice.actions.namespacesChanged(action.payload));
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
    }
  },
  reducer: {
    namespaces: namespacesSlice.reducer,
    preferences: preferencesSlice.reducer,
    meta: metaSlice.reducer,
    [namespaceApi.reducerPath]: namespaceApi.reducer,
    [flagsApi.reducerPath]: flagsApi.reducer,
    [segmentsApi.reducerPath]: segmentsApi.reducer,
    [rulesApi.reducerPath]: rulesApi.reducer,
    [rolloutsApi.reducerPath]: rolloutsApi.reducer,
    [tokensApi.reducerPath]: tokensApi.reducer
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .prepend(listenerMiddleware.middleware)
      .concat(
        namespaceApi.middleware,
        flagsApi.middleware,
        segmentsApi.middleware,
        rulesApi.middleware,
        rolloutsApi.middleware,
        tokensApi.middleware
      )
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
