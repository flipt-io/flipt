import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const eventKey = 'event';
interface IEventState {
  completedOnboarding: boolean;
  dismissedBanner_v1_49_0: boolean;
}

const initialState: IEventState = {
  completedOnboarding: false,
  dismissedBanner_v1_49_0: false
};

export const eventSlice = createSlice({
  name: 'event',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    },
    bannerDismissed: (state) => {
      state.dismissedBanner_v1_49_0 = true;
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
  (user) => user.dismissedBanner_v1_49_0
);

export default eventSlice.reducer;
