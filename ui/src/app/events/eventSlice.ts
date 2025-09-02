import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const eventKey = 'event';
interface IEventState {
  completedOnboarding: boolean;
  dismissedBanner_v2_release: boolean;
}

const initialState: IEventState = {
  completedOnboarding: false,
  dismissedBanner_v2_release: false
};

export const eventSlice = createSlice({
  name: 'event',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    },
    v2BannerDismissed: (state) => {
      state.dismissedBanner_v2_release = true;
    }
  }
});

export const { onboardingCompleted, v2BannerDismissed } = eventSlice.actions;

export const selectCompletedOnboarding = createSelector(
  [(state: RootState) => state.user],
  (user) => user.completedOnboarding
);

export const selectDismissedV2Banner = createSelector(
  [(state: RootState) => state.user],
  (user) => user.dismissedBanner_v2_release
);

export default eventSlice.reducer;
