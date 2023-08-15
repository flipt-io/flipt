import { IDistribution } from './Distribution';
import { ISegment, SegmentOperatorType } from './Segment';
import { IVariant } from './Variant';

export interface IVariantRollout {
  variant: IVariant;
  distribution: IDistribution;
}

export interface IEvaluatable {
  id: string;
  segment?: ISegment;
  segments: ISegment[];
  operator: SegmentOperatorType;
  rank: number;
  rollouts: IVariantRollout[];
  createdAt: string;
  updatedAt: string;
}
