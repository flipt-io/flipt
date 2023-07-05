import { IPageable } from './Pageable';

export enum RolloutType {
  UNKNOW_ROLLOUT_TYPE = 'Unknown',
  SEGMENT_ROLLOUT_TYPE = 'Segment',
  THRESHOLD_ROLLOUT_TYPE = 'Threshold'
}

export interface RolloutSegment {
  segmentKey: string;
  value: boolean;
}

export interface RolloutThreshold {
  percentage: number;
  value: boolean;
}

export interface IRolloutBase {
  type: RolloutType;
  rank: number;
  description?: string;
  rule: RolloutSegment | RolloutThreshold;
}

export interface IRollout extends IRolloutBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface IRolloutList extends IPageable {
  rollouts: IRollout[];
}
