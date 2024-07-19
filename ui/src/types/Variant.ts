import { IFilterable } from './Selectable';

export interface IVariantBase {
  key: string;
  name: string;
  description: string;
  attachment?: string;
}

export interface IVariant extends IVariantBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export type FilterableVariant = Pick<IVariant, 'id' | 'key' | 'name'> &
  IFilterable;
