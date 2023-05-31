import { configureStore } from '@reduxjs/toolkit';

import { namespacesSlice } from './app/namespaces/namespacesSlice';

export const store = configureStore({
  reducer: {
    namespaces: namespacesSlice.reducer
  }
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
