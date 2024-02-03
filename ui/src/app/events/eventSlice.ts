import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const eventKey = 'event';
interface IEventState {
  completedOnboarding: boolean;
}

const initialState: IEventState = {
  completedOnboarding: false
};

export const eventSlice = createSlice({
  name: 'event',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    }
  }
});

export const { onboardingCompleted } = eventSlice.actions;

export const selectCompletedOnboarding = createSelector(
  [(state: RootState) => state.user],
  (user) => user.completedOnboarding
);

export default eventSlice.reducer;
