import { FlagType, IFlag, MetadataFilter } from '~/types/Flag';

import { applyMetadataFilters } from './flagMetadataFilter';

const makeFlag = (key: string, metadata?: Record<string, any>): IFlag => ({
  key,
  name: key,
  type: FlagType.VARIANT,
  enabled: true,
  description: '',
  metadata
});

describe('applyMetadataFilters', () => {
  const flags = [
    makeFlag('flag-a', { team: 'backend', env: 'production' }),
    makeFlag('flag-b', { team: 'frontend', env: 'production' }),
    makeFlag('flag-c', { team: 'backend', env: 'staging' }),
    makeFlag('flag-d'), // no metadata
    makeFlag('flag-e', { count: 42, active: true })
  ];

  it('returns all flags when no filters are active', () => {
    expect(applyMetadataFilters(flags, [])).toEqual(flags);
  });

  it('filters by a single exact-match key-value pair', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'backend' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-c']);
  });

  it('filter is case-insensitive', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'BACKEND' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-c']);
  });

  it('filter is a substring match', () => {
    const filters: MetadataFilter[] = [{ key: 'env', value: 'prod' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-b']);
  });

  it('multiple filters use AND logic', () => {
    const filters: MetadataFilter[] = [
      { key: 'team', value: 'backend' },
      { key: 'env', value: 'production' }
    ];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a']);
  });

  it('excludes flags that do not have the filtered key', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'backend' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).not.toContain('flag-d');
  });

  it('matches numeric metadata values converted to string', () => {
    const filters: MetadataFilter[] = [{ key: 'count', value: '42' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-e']);
  });

  it('matches boolean metadata values converted to string', () => {
    const filters: MetadataFilter[] = [{ key: 'active', value: 'true' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-e']);
  });
});
