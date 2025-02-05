import { SegmentOperatorType } from '~/types/Segment';
import { IDistribution } from './Distribution';
import { IPageable } from './Pageable';

export interface IRuleBase {
  segments?: string[];
  segmentOperator?: SegmentOperatorType;
  rank: number;
  distributions: IDistribution[];
}

export interface IRule extends IRuleBase {
  id: string;
}

export interface IRuleList extends IPageable {
  rules: IRule[];
}
