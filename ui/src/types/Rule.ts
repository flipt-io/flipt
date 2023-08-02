import { OperatorType } from '~/types/Operator';
import { IDistribution } from './Distribution';
import { IPageable } from './Pageable';

export interface IRuleBase {
  segmentKey?: string;
  segmentKeys?: string[];
  segmentOperator?: OperatorType;
  rank: number;
}

export interface IRule extends IRuleBase {
  id: string;
  createdAt: string;
  updatedAt: string;
  distributions: IDistribution[];
}

export interface IRuleList extends IPageable {
  rules: IRule[];
}
