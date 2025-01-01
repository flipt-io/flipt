import { IPageable } from './Pageable';
import { IFilterable } from './Selectable';
import { IVariant } from './Variant';

export enum FlagType {
  VARIANT = 'VARIANT_FLAG_TYPE',
  BOOLEAN = 'BOOLEAN_FLAG_TYPE'
}

export function flagTypeToLabel(flagType: FlagType): string {
  switch (flagType) {
    case FlagType.BOOLEAN:
      return 'Boolean';
    case FlagType.VARIANT:
      return 'Variant';
    default:
      return 'Unknown';
  }
}

export interface IFlagMetadata {
  key: string;
  value: any;
  type: 'primitive' | 'object' | 'array';
  subtype?: 'string' | 'number' | 'boolean';
  isNew?: boolean;
}

export interface IFlagBase {
  key: string;
  type: FlagType;
  name: string;
  enabled: boolean;
  description: string;
  defaultVariant?: IVariant;
  metadata?: Record<string, any>;
}

export interface IFlag extends IFlagBase {
  createdAt: string;
  updatedAt: string;
  variants?: IVariant[];
}

export interface IFlagList extends IPageable {
  flags: IFlag[];
}

export type FilterableFlag = IFlag & IFilterable;
