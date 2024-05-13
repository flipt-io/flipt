import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const eventKey = 'event';
interface IEventState {
  completedOnboarding: boolean;
  dismissedBanner: boolean;
}

const initialState: IEventState = {
  completedOnboarding: false,
  dismissedBanner: false
};

export const eventSlice = createSlice({
  name: 'event',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    },
    bannerDismissed: (state) => {
      state.dismissedBanner = true;
    }
  }
});

export const { onboardingCompleted, bannerDismissed } = eventSlice.actions;

export const selectCompletedOnboarding = createSelector(
  [(state: RootState) => state.user],
  (user) => user.completedOnboarding
);

export const selectDismissedBanner = createSelector(
  [(state: RootState) => state.user],
  (user) => user.dismissedBanner
);

export default eventSlice.reducer;
