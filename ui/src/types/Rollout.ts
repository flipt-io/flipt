import { IPageable } from './Pageable';
import { SegmentOperatorType } from './Segment';

export enum RolloutType {
  SEGMENT = 'SEGMENT_ROLLOUT_TYPE',
  THRESHOLD = 'THRESHOLD_ROLLOUT_TYPE'
}

export function rolloutTypeToLabel(rolloutType: RolloutType): string {
  switch (rolloutType) {
    case RolloutType.SEGMENT:
      return 'Segment';
    case RolloutType.THRESHOLD:
      return 'Threshold';
    default:
      return 'Unknown';
  }
}

export interface IRolloutRuleSegment {
  segmentOperator?: SegmentOperatorType;
  segmentKey?: string;
  segmentKeys?: string[];
  value: boolean;
}

export interface IRolloutRuleThreshold {
  percentage: number;
  value: boolean;
}

export interface IRolloutBase {
  type: RolloutType;
  rank: number;
  description?: string;
  threshold?: IRolloutRuleThreshold;
  segment?: IRolloutRuleSegment;
}

export interface IRollout extends IRolloutBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface IRolloutList extends IPageable {
  rules: IRollout[];
}
