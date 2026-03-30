import { areDifferent, diffFlag, diffSegment, driftFields, stableValue, toResourceType, toStatus } from './compare';

describe('stableValue', () => {
  it('handles primitives', () => {
    expect(stableValue(42)).toBe('42');
    expect(stableValue('hello')).toBe('"hello"');
    expect(stableValue(null)).toBe('null');
    expect(stableValue(true)).toBe('true');
  });

  it('handles arrays', () => {
    expect(stableValue([1, 2, 3])).toBe('[1,2,3]');
  });

  it('sorts object keys for stable comparison', () => {
    const a = stableValue({ b: 2, a: 1 });
    const b = stableValue({ a: 1, b: 2 });
    expect(a).toBe(b);
  });

  it('handles nested objects', () => {
    const result = stableValue({ z: { b: 2, a: 1 }, y: [3, 4] });
    expect(result).toContain('a:1');
    expect(result).toContain('b:2');
  });
});

describe('areDifferent', () => {
  it('returns false for identical values', () => {
    expect(areDifferent(42, 42)).toBe(false);
    expect(areDifferent({ a: 1 }, { a: 1 })).toBe(false);
  });

  it('returns true for different values', () => {
    expect(areDifferent(42, 43)).toBe(true);
    expect(areDifferent({ a: 1 }, { a: 2 })).toBe(true);
  });

  it('ignores object key ordering', () => {
    expect(areDifferent({ b: 2, a: 1 }, { a: 1, b: 2 })).toBe(false);
  });
});

describe('diffFlag', () => {
  const base = {
    enabled: true,
    defaultVariant: 'v1',
    variants: [{ key: 'v1' }],
    rules: [],
    rollouts: []
  };

  it('returns empty for identical flags', () => {
    expect(diffFlag(base, { ...base })).toEqual([]);
  });

  it('detects enabled difference', () => {
    expect(diffFlag(base, { ...base, enabled: false })).toContain('enabled');
  });

  it('detects variant differences', () => {
    expect(diffFlag(base, { ...base, variants: [{ key: 'v2' }] })).toContain('variants');
  });

  it('detects multiple differences', () => {
    const diff = diffFlag(base, { ...base, enabled: false, rules: [{ id: '1' }] });
    expect(diff).toContain('enabled');
    expect(diff).toContain('rules');
  });
});

describe('driftFields', () => {
  const base = {
    enabled: true,
    defaultVariant: 'v1',
    variants: [{ key: 'v1' }],
    rules: [],
    rollouts: []
  };

  it('excludes enabled/disabled from drift', () => {
    expect(driftFields(base, { ...base, enabled: false })).toEqual([]);
  });

  it('includes structural differences', () => {
    const drift = driftFields(base, { ...base, variants: [{ key: 'v2' }] });
    expect(drift).toContain('variants');
  });

  it('detects default variant drift', () => {
    const drift = driftFields(base, { ...base, defaultVariant: 'v2' });
    expect(drift).toContain('default variant');
  });
});

describe('diffSegment', () => {
  const base = { matchType: 'ALL', constraints: [{ type: 'eq' }] };

  it('returns empty for identical segments', () => {
    expect(diffSegment(base, { ...base })).toEqual([]);
  });

  it('detects constraint differences', () => {
    expect(diffSegment(base, { ...base, constraints: [] })).toContain('constraints');
  });
});

describe('toStatus', () => {
  it('maps known statuses', () => {
    expect(toStatus('COMPARE_STATUS_DIFFERENT')).toBe('different');
    expect(toStatus('COMPARE_STATUS_SOURCE_ONLY')).toBe('source_only');
    expect(toStatus('COMPARE_STATUS_TARGET_ONLY')).toBe('target_only');
    expect(toStatus('COMPARE_STATUS_IDENTICAL')).toBe('same');
  });

  it('defaults to same for unknown values', () => {
    expect(toStatus('UNKNOWN')).toBe('same');
  });
});

describe('toResourceType', () => {
  it('maps known types', () => {
    expect(toResourceType('flipt.core.Flag')).toBe('flag');
    expect(toResourceType('flipt.core.Segment')).toBe('segment');
  });

  it('returns null for unknown types', () => {
    expect(toResourceType('unknown')).toBeNull();
  });
});
