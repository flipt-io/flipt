import { useCallback, useEffect, useRef } from 'react';

import {
  StreamEvent,
  clearEvents,
  connectionClosed,
  connectionError,
  connectionEstablished,
  eventReceived
} from '~/app/flags/streamingApi';

import { useAppDispatch } from '~/data/hooks/store';

interface UseChangesStreamOptions {
  environmentKey: string;
  namespaceKey: string;
  enabled?: boolean;
}

export function useChangesStream({
  environmentKey,
  namespaceKey,
  enabled = true
}: UseChangesStreamOptions) {
  const dispatch = useAppDispatch();
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
    null
  );

  const connect = useCallback(() => {
    if (!enabled) return;

    const url = `/client/v2/environments/${environmentKey}/namespaces/${namespaceKey}/stream`;

    try {
      const eventSource = new EventSource(url);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        dispatch(connectionEstablished());
      };

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          const streamEvent: StreamEvent = {
            type: data.type || 'message',
            timestamp: new Date().toISOString(),
            data
          };
          dispatch(eventReceived(streamEvent));
        } catch {
          const streamEvent: StreamEvent = {
            type: 'message',
            timestamp: new Date().toISOString(),
            data: { raw: event.data }
          };
          dispatch(eventReceived(streamEvent));
        }
      };

      eventSource.onerror = () => {
        dispatch(connectionError('Connection failed'));
        eventSource.close();
        eventSourceRef.current = null;

        if (enabled) {
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, 5000);
        }
      };
    } catch (err) {
      dispatch(
        connectionError(err instanceof Error ? err.message : 'Unknown error')
      );
    }
  }, [dispatch, enabled, environmentKey, namespaceKey]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    dispatch(connectionClosed());
  }, [dispatch]);

  const clear = useCallback(() => {
    dispatch(clearEvents());
  }, [dispatch]);

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    connect,
    disconnect,
    clear
  };
}
