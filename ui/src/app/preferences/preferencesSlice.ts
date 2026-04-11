/* eslint-disable @typescript-eslint/no-use-before-define */
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '~/store';
import { Theme, Timezone, Sidebar } from '~/types/Preferences';
import { fetchInfoAsync } from '~/app/meta/metaSlice';

export const preferencesKey = 'preferences';

interface IPreferencesState {
  theme: Theme;
  timezone: Timezone;
  sidebar: Sidebar;
  /** When true, the UI loads which flags reference each segment (extra API calls). */
  loadSegmentFlagReferences?: boolean;
}

const initialState: IPreferencesState = {
  theme: Theme.SYSTEM,
  timezone: Timezone.LOCAL,
  sidebar: Sidebar.OPEN,
  loadSegmentFlagReferences: false
};

export const preferencesSlice = createSlice({
  name: 'preferences',
  initialState,
  reducers: {
    themeChanged: (state, action) => {
      state.theme = action.payload;
    },
    timezoneChanged: (state, action) => {
      state.timezone = action.payload;
    },
    sidebarChanged: (state, action: PayloadAction<boolean>) => {
      state.sidebar = action.payload ? Sidebar.OPEN : Sidebar.CLOSE;
    },
    loadSegmentFlagReferencesChanged: (
      state,
      action: PayloadAction<boolean>
    ) => {
      state.loadSegmentFlagReferences = action.payload;
    }
  },
  extraReducers(builder) {
    builder.addCase(fetchInfoAsync.fulfilled, (state, action) => {
      const currentPreference = JSON.parse(
        localStorage.getItem(preferencesKey) || '{}'
      ) as IPreferencesState;
      // If there isn't currently a set theme, set to the default theme
      if (!currentPreference.theme) {
        state.theme = action.payload.ui?.theme || 'system';
      }

      if (!currentPreference.timezone) {
        state.timezone = Timezone.LOCAL;
      }
      let sidebar = Sidebar.OPEN;
      if (
        currentPreference.sidebar &&
        Object.values(Sidebar).includes(currentPreference.sidebar)
      ) {
        sidebar = currentPreference.sidebar;
      }
      state.sidebar = sidebar;

      if (typeof currentPreference.loadSegmentFlagReferences === 'boolean') {
        state.loadSegmentFlagReferences =
          currentPreference.loadSegmentFlagReferences;
      }
    });
  }
});

export const {
  themeChanged,
  timezoneChanged,
  sidebarChanged,
  loadSegmentFlagReferencesChanged
} = preferencesSlice.actions;

export const selectTheme = (state: RootState) => state.preferences.theme;
export const selectTimezone = (state: RootState) => state.preferences.timezone;
export const selectSidebar = (state: RootState) =>
  state.preferences.sidebar === Sidebar.OPEN;
/** Off by default; must be explicitly enabled in Preferences. */
export const selectLoadSegmentFlagReferences = (state: RootState) =>
  state.preferences.loadSegmentFlagReferences === true;

export default preferencesSlice.reducer;