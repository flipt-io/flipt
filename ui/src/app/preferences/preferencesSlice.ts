/* eslint-disable @typescript-eslint/no-use-before-define */
import { createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';
import { Theme, Timezone } from '~/types/Preferences';
import { fetchInfoAsync } from '~/app/meta/metaSlice';

export const preferencesKey = 'preferences';

interface IPreferencesState {
  theme: Theme;
  timezone: Timezone;
}

const initialState: IPreferencesState = {
  theme: Theme.SYSTEM,
  timezone: Timezone.LOCAL
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
    });
  }
});

export const { themeChanged, timezoneChanged } = preferencesSlice.actions;

export const selectTheme = (state: RootState) => state.preferences.theme;
export const selectTimezone = (state: RootState) => state.preferences.timezone;

export default preferencesSlice.reducer;
