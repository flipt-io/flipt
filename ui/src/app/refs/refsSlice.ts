import { createSelector, createSlice } from '@reduxjs/toolkit';
import { RootState } from '~/store';

export const refsKey = 'refs';
interface IRefState {
  currentRef?: string;
}

const initialState: IRefState = {
  currentRef: undefined
};

export const refsSlice = createSlice({
  name: 'refs',
  initialState,
  reducers: {
    currentRefChanged: (state, action) => {
      const ref = action.payload;
      state.currentRef = ref;
    }
  }
});

export const { currentRefChanged } = refsSlice.actions;

export const selectCurrentRef = createSelector(
  [(state: RootState) => state.refs],
  (refs) => refs.currentRef
);

export default refsSlice.reducer;
