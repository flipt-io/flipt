import { IConstraint } from './Constraint';
import { IPageable } from './Pageable';

export interface ISegmentBase {
  key: string;
  name: string;
  description: string;
  matchType: SegmentMatchType;
}

export interface ISegment extends ISegmentBase {
  createdAt: string;
  updatedAt: string;
  constraints?: IConstraint[];
}

export enum SegmentMatchType {
  ALL_MATCH_TYPE = 'All',
  ANY_MATCH_TYPE = 'Any'
}

export interface ISegmentList extends IPageable {
  segments: ISegment[];
}
