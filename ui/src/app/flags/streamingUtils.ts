export const shouldInvalidateFromStreamEvent = (
  eventData: { type?: string; etag?: string } | null
) => eventData?.type !== 'error';
