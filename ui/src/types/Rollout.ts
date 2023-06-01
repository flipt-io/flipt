import { IPageable } from './Pageable';

export enum RolloutType {
  PERCENT_ROLLOUT_TYPE = 'percent',
  SEGMENT_ROLLOUT_TYPE = 'segment'
}

export interface IRolloutBase {
  flagKey: string;
  description: string;
  type: RolloutType;
  rank: number;
  rule: IRolloutPercent | IRolloutSegment;
}

export interface IRollout extends IRolloutBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface IRolloutList extends IPageable {
  rollouts: IRollout[];
}

export interface IRolloutPercent {
  percentage: number;
  value: boolean;
}

export interface IRolloutSegment {
  segmentId: string;
  value: boolean;
}
