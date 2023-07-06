import { IPageable } from './Pageable';

export enum RolloutType {
  SEGMENT_ROLLOUT_TYPE = 'Segment',
  THRESHOLD_ROLLOUT_TYPE = 'Threshold'
}

export interface IRolloutRuleSegment {
  segmentKey: string;
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
  rollouts: IRollout[];
}
