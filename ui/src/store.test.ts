import { shouldInvalidateFromStreamEvent } from '~/app/flags/streamingUtils';

describe('shouldInvalidateFromStreamEvent', () => {
  it('invalidates for refetchEvaluation events', () => {
    expect(shouldInvalidateFromStreamEvent({ type: 'refetchEvaluation' })).toBe(
      true
    );
  });

  it('invalidates for stream payloads without a type', () => {
    expect(shouldInvalidateFromStreamEvent({ etag: 'digest' })).toBe(true);
  });

  it('does not invalidate for error events', () => {
    expect(shouldInvalidateFromStreamEvent({ type: 'error' })).toBe(false);
  });
});
