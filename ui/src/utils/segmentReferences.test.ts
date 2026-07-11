import { FlagType, IFlag } from '~/types/Flag';
import { RolloutType } from '~/types/Rollout';

import {
  flagReferencesSegment,
  flagsReferencingSegment
} from './segmentReferences';

const baseFlag = (key: string): IFlag => ({
  key,
  name: key,
  type: FlagType.BOOLEAN,
  enabled: true,
  description: ''
});

describe('flagReferencesSegment', () => {
  it('returns true when a flag rule references the segment', () => {
    const flag = {
      ...baseFlag('checkout'),
      rules: [
        {
          segments: ['beta-users'],
          distributions: []
        }
      ]
    };

    expect(flagReferencesSegment(flag, 'beta-users')).toBe(true);
  });

  it('returns true when a rollout references the segment', () => {
    const flag = {
      ...baseFlag('pricing'),
      rollouts: [
        {
          type: RolloutType.SEGMENT,
          segment: {
            segments: ['internal-users'],
            value: true
          }
        }
      ]
    };

    expect(flagReferencesSegment(flag, 'internal-users')).toBe(true);
  });

  it('returns false when the segment is not referenced', () => {
    const flag = {
      ...baseFlag('search'),
      rules: [
        {
          segments: ['other-segment'],
          distributions: []
        }
      ],
      rollouts: [
        {
          type: RolloutType.SEGMENT,
          segment: {
            segments: ['another-segment'],
            value: true
          }
        }
      ]
    };

    expect(flagReferencesSegment(flag, 'beta-users')).toBe(false);
  });
});

describe('flagsReferencingSegment', () => {
  it('filters flags to only those referencing the segment', () => {
    const flags = [
      {
        ...baseFlag('checkout'),
        rules: [
          {
            segments: ['beta-users'],
            distributions: []
          }
        ]
      },
      baseFlag('pricing'),
      {
        ...baseFlag('search'),
        rollouts: [
          {
            type: RolloutType.SEGMENT,
            segment: {
              segments: ['beta-users'],
              value: true
            }
          }
        ]
      }
    ];

    expect(
      flagsReferencingSegment(flags, 'beta-users').map((f) => f.key)
    ).toEqual(['checkout', 'search']);
  });
});
