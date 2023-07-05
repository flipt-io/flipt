import { IPageable } from './Pageable';
import { IVariant } from './Variant';

export enum FlagType {
  VARIANT_FLAG_TYPE = 'Variant',
  BOOLEAN_FLAG_TYPE = 'Boolean'
}

export interface IFlagBase {
  key: string;
  type: FlagType;
  name: string;
  enabled: boolean;
  description: string;
}

export interface IFlag extends IFlagBase {
  createdAt: string;
  updatedAt: string;
  variants?: IVariant[];
}

export interface IFlagList extends IPageable {
  flags: IFlag[];
}
