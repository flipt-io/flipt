/**
 * Stream payloads without an explicit `type` can still represent data updates.
 * We only skip invalidation for explicit error events.
 */
export const shouldInvalidateFromStreamEvent = (
  eventData: { type?: string; etag?: string } | null
) => eventData !== null && eventData.type !== 'error';
