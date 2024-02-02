import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const userKey = 'user';
interface IUserState {
  completedOnboarding: boolean;
}

const initialState: IUserState = {
  completedOnboarding: false
};

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    onboardingCompleted: (state) => {
      state.completedOnboarding = true;
    }
  }
});

export const { onboardingCompleted } = userSlice.actions;

export const selectCompletedOnboarding = createSelector(
  [(state: RootState) => state.user.completedOnboarding],
  (completedOnboarding) => completedOnboarding
);

export default userSlice.reducer;
