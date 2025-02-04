import {
  configureStore,
  createListenerMiddleware,
  isAnyOf
} from '@reduxjs/toolkit';

import { authProvidersApi } from '~/app/auth/authApi';
import {
  namespaceApi,
  namespaceKey,
  namespacesSlice
} from '~/app/namespaces/namespacesApi';
import { analyticsApi } from './app/flags/analyticsApi';
import { flagsApi, flagsTableSlice } from './app/flags/flagsApi';
import { rolloutTag, rolloutsApi } from './app/flags/rolloutsApi';
import { ruleTag, rulesApi } from './app/flags/rulesApi';
import { metaSlice } from './app/meta/metaSlice';
import {
  preferencesKey,
  preferencesSlice
} from './app/preferences/preferencesSlice';
import {
  segmentTag,
  segmentsApi,
  segmentsTableSlice
} from './app/segments/segmentsApi';
import { tokensApi } from './app/tokens/tokensApi';
import { LoadingStatus } from './types/Meta';
import { eventSlice, eventKey } from './app/events/eventSlice';
import {
  environmentKey,
  environmentsApi,
  environmentsSlice
} from './app/environments/environmentsApi';

const listenerMiddleware = createListenerMiddleware();

listenerMiddleware.startListening({
  matcher: isAnyOf(
    eventSlice.actions.onboardingCompleted,
    eventSlice.actions.bannerDismissed
  ),
  effect: (_action, api) => {
    // save to local storage
    localStorage.setItem(
      eventKey,
      JSON.stringify((api.getState() as RootState).user)
    );
  }
});

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

// clean segments cache for deleted namespace
listenerMiddleware.startListening({
  matcher: namespaceApi.endpoints.deleteNamespace.matchFulfilled,
  effect: (action, api) => {
    api.dispatch(
      segmentsApi.util.invalidateTags([
        segmentTag(action.meta.arg.originalArgs)
      ])
    );
  }
});

// clean rollouts and rules cache for deleted flag
listenerMiddleware.startListening({
  matcher: flagsApi.endpoints.deleteFlag.matchFulfilled,
  effect: (action, api) => {
    const arg = action.meta.arg.originalArgs;
    api.dispatch(rolloutsApi.util.invalidateTags([rolloutTag(arg)]));
    api.dispatch(rulesApi.util.invalidateTags([ruleTag(arg)]));
  }
});

const userState = JSON.parse(localStorage.getItem(eventKey) || '{}');

const preferencesState = JSON.parse(
  localStorage.getItem(preferencesKey) || '{}'
);

const currentEnvironment = localStorage.getItem(environmentKey) || 'default';
const currentNamespace = localStorage.getItem(namespaceKey) || 'default';

export const store = configureStore({
  preloadedState: {
    user: userState,
    preferences: preferencesState,
    environments: {
      environments: {},
      status: LoadingStatus.IDLE,
      currentEnvironment,
      error: undefined
    },
    namespaces: {
      namespaces: {},
      status: LoadingStatus.IDLE,
      currentNamespace,
      error: undefined
    },
    flagsTable: {
      sorting: []
    },
    segmentsTable: {
      sorting: []
    }
  },
  reducer: {
    flagsTable: flagsTableSlice.reducer,
    segmentsTable: segmentsTableSlice.reducer,
    user: eventSlice.reducer,
    preferences: preferencesSlice.reducer,
    environments: environmentsSlice.reducer,
    namespaces: namespacesSlice.reducer,
    meta: metaSlice.reducer,
    [environmentsApi.reducerPath]: environmentsApi.reducer,
    [namespaceApi.reducerPath]: namespaceApi.reducer,
    [flagsApi.reducerPath]: flagsApi.reducer,
    [segmentsApi.reducerPath]: segmentsApi.reducer,
    [rulesApi.reducerPath]: rulesApi.reducer,
    [rolloutsApi.reducerPath]: rolloutsApi.reducer,
    [tokensApi.reducerPath]: tokensApi.reducer,
    [authProvidersApi.reducerPath]: authProvidersApi.reducer,
    [analyticsApi.reducerPath]: analyticsApi.reducer
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .prepend(listenerMiddleware.middleware)
      .concat(
        environmentsApi.middleware,
        namespaceApi.middleware,
        flagsApi.middleware,
        segmentsApi.middleware,
        rulesApi.middleware,
        rolloutsApi.middleware,
        tokensApi.middleware,
        authProvidersApi.middleware,
        analyticsApi.middleware
      )
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
