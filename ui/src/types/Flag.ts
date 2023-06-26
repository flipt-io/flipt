import { IPageable } from './Pageable';
import { IVariant } from './Variant';

export interface IFlagBase {
  key: string;
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
