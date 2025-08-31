import { createSelector, createSlice } from '@reduxjs/toolkit';

import { RootState } from '~/store';

export const eventKey = 'event';
interface IEventState {
  completedOnboarding: boolean;
  dismissedProBanner: boolean;
}

const initialState: IEventState = {
  completedOnboarding: false,
  dismissedProBanner: false
};

export const eventSlice = createSlice({
  name: 'event',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    },
    proBannerDismissed: (state) => {
      state.dismissedProBanner = true;
    }
  }
});

export const { onboardingCompleted, proBannerDismissed } = eventSlice.actions;

export const selectCompletedOnboarding = createSelector(
  [(state: RootState) => state.user],
  (user) => user.completedOnboarding
);

export const selectDismissedProBanner = createSelector(
  [(state: RootState) => state.user],
  (user) => user.dismissedProBanner
);

export default eventSlice.reducer;
