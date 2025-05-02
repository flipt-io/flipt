import { SegmentOperatorType } from '~/types/Segment';

import { IDistribution } from './Distribution';
import { IPageable } from './Pageable';

export interface IRule {
  id?: string; // for dnd-drag-and-drop
  segments?: string[];
  segmentOperator?: SegmentOperatorType;
  rank?: number;
  distributions: IDistribution[];
}

export interface IRuleList extends IPageable {
  rules: IRule[];
}
