import {
  configureStore,
  createListenerMiddleware,
  isAnyOf
} from '@reduxjs/toolkit';

import { authProvidersApi } from '~/app/auth/authApi';
import {
  environmentKey,
  environmentsApi,
  environmentsSlice
} from '~/app/environments/environmentsApi';
import { eventKey, eventSlice } from '~/app/events/eventSlice';
import { analyticsApi } from '~/app/flags/analyticsApi';
import { flagsApi, flagsTableSlice } from '~/app/flags/flagsApi';
import { metaSlice } from '~/app/meta/metaSlice';
import {
  namespaceApi,
  namespaceKey,
  namespacesSlice
} from '~/app/namespaces/namespacesApi';
import {
  preferencesKey,
  preferencesSlice
} from '~/app/preferences/preferencesSlice';
import { segmentsApi, segmentsTableSlice } from '~/app/segments/segmentsApi';

import { LoadingStatus } from '~/types/Meta';

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
  matcher: isAnyOf(environmentsSlice.actions.currentEnvironmentChanged),
  effect: (_action, api) => {
    // save to local storage
    localStorage.setItem(
      environmentKey,
      (api.getState() as RootState).environments.currentEnvironment
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
 * Each time the the app is loaded or environments changed by user this
 * listener will propagate the latest environments to the environmentsSlice.
 */
listenerMiddleware.startListening({
  matcher: environmentsApi.endpoints.listEnvironments.matchFulfilled,
  effect: (action, api) => {
    api.dispatch(environmentsSlice.actions.environmentsChanged(action.payload));
  }
});

/*
 * Each time the the app is loaded or namespaces changed by user this
 * listener will propagate the latest namespaces to the namespacesSlice.
 */
listenerMiddleware.startListening({
  matcher: namespaceApi.endpoints.listNamespaces.matchFulfilled,
  effect: (action, api) => {
    api.dispatch(namespacesSlice.actions.namespacesChanged(action.payload));
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
        authProvidersApi.middleware,
        analyticsApi.middleware
      )
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
