/**
 * @jest-environment jsdom
 * @jest-environment-options {"url": "https://test/"}
 */
import { flagsApi } from '~/app/flags/flagsApi';
import { eventReceived } from '~/app/flags/streamingApi';
import {
  currentNamespaceChanged,
  namespaceKey
} from '~/app/namespaces/namespacesApi';
import { segmentsApi } from '~/app/segments/segmentsApi';

import { store } from '~/store';

function spyOnApiUtil<
  O extends Record<string, (...args: unknown[]) => unknown>,
  M extends keyof O
>(obj: O, method: M): { spy: jest.Mock; restore: () => void } {
  const orig = obj[method];

  if (typeof orig !== 'function') {
    throw new Error(
      `spyOnApiUtil: "${String(method)}" is not a function on the given object`
    );
  }

  const spy = jest
    .fn()
    .mockImplementation((...args: unknown[]) => orig(...args));

  for (const key of Object.getOwnPropertyNames(orig)) {
    const descriptor = Object.getOwnPropertyDescriptor(orig, key);
    if (descriptor && descriptor.writable) {
      (spy as unknown as Record<string, unknown>)[key] = (
        orig as unknown as Record<string, unknown>
      )[key];
    }
  }

  obj[method] = spy as unknown as O[M];
  return {
    spy,
    restore: () => {
      obj[method] = orig;
    }
  };
}

beforeEach(() => {
  localStorage.clear();
  jest.restoreAllMocks();
});

describe('namespace localStorage', () => {
  it('stores namespace key in localStorage when set to a truthy value', () => {
    store.dispatch(currentNamespaceChanged('production'));
    expect(localStorage.getItem(namespaceKey)).toBe('production');
  });

  it('removes namespace key from localStorage when cleared', () => {
    localStorage.setItem(namespaceKey, 'production');
    store.dispatch(currentNamespaceChanged(''));
    expect(localStorage.getItem(namespaceKey)).toBeNull();
  });
});

describe('SSE cache invalidation', () => {
  let flagSpy: jest.Mock;
  let segmentSpy: jest.Mock;
  let restoreFlag: () => void;
  let restoreSegment: () => void;

  beforeEach(() => {
    const f = spyOnApiUtil(
      flagsApi.util as Record<string, (...args: unknown[]) => unknown>,
      'invalidateTags'
    );
    const s = spyOnApiUtil(
      segmentsApi.util as Record<string, (...args: unknown[]) => unknown>,
      'invalidateTags'
    );
    flagSpy = f.spy;
    segmentSpy = s.spy;
    restoreFlag = f.restore;
    restoreSegment = s.restore;
  });

  afterEach(() => {
    restoreFlag();
    restoreSegment();
  });

  it('falls back to default/default when envKey and nsKey are not in event data', () => {
    store.dispatch(
      eventReceived({
        type: 'refetchEvaluation',
        timestamp: new Date().toISOString(),
        data: { type: 'refetchEvaluation' }
      })
    );

    expect(flagSpy).toHaveBeenCalledWith([
      { type: 'Flag', id: 'default/default' }
    ]);
    expect(segmentSpy).toHaveBeenCalledWith([
      { type: 'Segment', id: 'default/default' }
    ]);
  });

  it('uses envKey and nsKey from event data when provided', () => {
    store.dispatch(
      eventReceived({
        type: 'refetchEvaluation',
        timestamp: new Date().toISOString(),
        data: {
          type: 'refetchEvaluation',
          envKey: 'staging',
          nsKey: 'email'
        }
      })
    );

    expect(flagSpy).toHaveBeenCalledWith([
      { type: 'Flag', id: 'staging/email' }
    ]);
    expect(segmentSpy).toHaveBeenCalledWith([
      { type: 'Segment', id: 'staging/email' }
    ]);
  });
});
