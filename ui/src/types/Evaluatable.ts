import { IDistribution } from './Distribution';
import { ISegment } from './Segment';
import { IVariant } from './Variant';

export interface IVariantRollout {
  variant: IVariant;
  distribution: IDistribution;
}

export interface IEvaluatable {
  id: string;
  segment: ISegment;
  rank: number;
  rollouts: IVariantRollout[];
  createdAt: string;
  updatedAt: string;
}
