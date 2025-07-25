/* eslint-disable @typescript-eslint/no-use-before-define */
import { PayloadAction, createSlice } from '@reduxjs/toolkit';

import { fetchInfoAsync } from '~/app/meta/metaSlice';

import { Sidebar, Theme, Timezone } from '~/types/Preferences';

import { RootState } from '~/store';

export const preferencesKey = 'preferences';

interface PreferencesState {
  theme: Theme;
  timezone: Timezone;
  lastSaved: number | null;
  sidebar: Sidebar;
}

const getInitialState = (): PreferencesState => {
  let theme = Theme.SYSTEM;
  let timezone = Timezone.LOCAL;
  let sidebar = Sidebar.OPEN;

  try {
    const storedPreferences = localStorage.getItem(preferencesKey);
    if (storedPreferences) {
      const preferences = JSON.parse(
        storedPreferences
      ) as Partial<PreferencesState>;

      if (
        preferences.theme &&
        Object.values(Theme).includes(preferences.theme)
      ) {
        theme = preferences.theme;
      }

      if (
        preferences.timezone &&
        Object.values(Timezone).includes(preferences.timezone)
      ) {
        timezone = preferences.timezone;
      }

      if (
        preferences.sidebar &&
        Object.values(Sidebar).includes(preferences.sidebar)
      ) {
        sidebar = preferences.sidebar;
      }
    }
  } catch (e) {
    // localStorage is disabled or not available, ignore
  }

  return {
    theme,
    timezone,
    sidebar,
    lastSaved: null
  };
};

const initialState: PreferencesState = getInitialState();

// Helper function to store preferences in localStorage
const savePreferences = (state: PreferencesState) => {
  try {
    localStorage.setItem(
      preferencesKey,
      JSON.stringify({
        theme: state.theme,
        timezone: state.timezone,
        sidebar: state.sidebar
      })
    );
    state.lastSaved = Date.now();
  } catch (e) {
    // localStorage is disabled or not available, ignore
  }
};

export const preferencesSlice = createSlice({
  name: 'preferences',
  initialState,
  reducers: {
    themeChanged(state, action: PayloadAction<Theme>) {
      state.theme = action.payload;
      savePreferences(state);
    },
    timezoneChanged(state, action: PayloadAction<Timezone>) {
      state.timezone = action.payload;
      savePreferences(state);
    },
    // Reset the lastSaved timestamp - used for debouncing notifications
    resetLastSaved(state) {
      state.lastSaved = null;
    },
    sidebarChanged(state, action: PayloadAction<boolean>) {
      state.sidebar = action.payload ? Sidebar.OPEN : Sidebar.CLOSE;
      savePreferences(state);
    }
  },
  extraReducers(builder) {
    builder.addCase(fetchInfoAsync.fulfilled, (state, action) => {
      const storedPreferences = localStorage.getItem(preferencesKey);
      const currentPreference = storedPreferences
        ? (JSON.parse(storedPreferences) as Partial<PreferencesState>)
        : {};

      // If there isn't currently a set theme, set to the default theme
      if (!currentPreference.theme) {
        state.theme = action.payload.ui.theme ?? Theme.SYSTEM;
      }

      if (!currentPreference.timezone) {
        state.timezone = Timezone.LOCAL;
      }

      if (currentPreference.sidebar === undefined) {
        currentPreference.sidebar = Sidebar.OPEN;
      }

      // Save the updated state
      savePreferences(state);
    });
  }
});

export const { themeChanged, timezoneChanged, resetLastSaved, sidebarChanged } =
  preferencesSlice.actions;

export const selectPreferences = (state: RootState) => state.preferences;

export const selectTheme = (state: RootState) => state.preferences.theme;

export const selectTimezone = (state: RootState) => state.preferences.timezone;
export const selectSidebar = (state: RootState) => {
  return state.preferences.sidebar == Sidebar.OPEN;
};

export const selectLastSaved = (state: RootState) =>
  state.preferences.lastSaved;

export default preferencesSlice.reducer;
