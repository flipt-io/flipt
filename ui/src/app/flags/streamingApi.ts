import { PayloadAction, createSlice } from '@reduxjs/toolkit';

export interface StreamEvent {
  type: string;
  timestamp: string;
  data?: Record<string, unknown>;
}

interface StreamingState {
  connected: boolean;
  events: StreamEvent[];
  lastEvent: StreamEvent | null;
  error: string | null;
}

const initialState: StreamingState = {
  connected: false,
  events: [],
  lastEvent: null,
  error: null
};

export const streamingSlice = createSlice({
  name: 'streaming',
  initialState,
  reducers: {
    connectionEstablished: (state) => {
      state.connected = true;
      state.error = null;
    },
    connectionClosed: (state) => {
      state.connected = false;
    },
    connectionError: (state, action: PayloadAction<string>) => {
      state.connected = false;
      state.error = action.payload;
    },
    eventReceived: (state, action: PayloadAction<StreamEvent>) => {
      state.lastEvent = action.payload;
      state.events.push(action.payload);
      if (state.events.length > 100) {
        state.events = state.events.slice(-100);
      }
    },
    clearEvents: (state) => {
      state.events = [];
      state.lastEvent = null;
    }
  }
});

export const {
  connectionEstablished,
  connectionClosed,
  connectionError,
  eventReceived,
  clearEvents
} = streamingSlice.actions;

export const streamingReducer = streamingSlice.reducer;

export const streamingKey = 'streaming';
