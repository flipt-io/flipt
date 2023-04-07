import { IDistribution } from './Distribution';
import { IFlag } from './Flag';
import { ISegment } from './Segment';
import { IVariant } from './Variant';

export interface IRollout {
  variant: IVariant;
  distribution: IDistribution;
}

export interface IEvaluatable {
  id: string;
  flag: IFlag;
  segment: ISegment;
  rank: number;
  rollouts: IRollout[];
  createdAt: string;
  updatedAt: string;
}
